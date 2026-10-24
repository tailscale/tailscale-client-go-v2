package coverage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"tailscale.com/client/tailscale/v2/tools/internal/openapi"
	"tailscale.com/client/tailscale/v2/tools/internal/repoanalysis"
)

type Report struct {
	SpecPath string
	RepoRoot string
	Endpoint EndpointReport
	Model    ModelReport
}

type EndpointReport struct {
	TotalSpecOperations        int
	TotalImplementedOperations int
	CoveredOperations          int
	Missing                    []openapi.Operation
	Extra                      []repoanalysis.Operation
	ByTag                      []TagCoverage
}

type TagCoverage struct {
	Tag     string
	Total   int
	Covered int
	Missing int
}

type ModelReport struct {
	TotalSpecModels     int
	CoveredModels       int
	TotalSpecProperties int
	CoveredProperties   int
	Matches             []ModelMatch
	MissingModels       []openapi.Model
	ExtraModels         []repoanalysis.Model
	ModelsWithGaps      []ModelMatch
}

type ModelMatch struct {
	Spec              openapi.Model
	Repo              repoanalysis.Model
	MatchType         string
	CoveredProperties int
	MissingProperties []string
	ExtraProperties   []repoanalysis.Field
}

func Build(spec *openapi.Spec, repo *repoanalysis.Analyzer, specPath, repoRoot string) (*Report, error) {
	endpointReport, err := buildEndpointReport(spec, repo)
	if err != nil {
		return nil, err
	}

	modelReport, err := buildModelReport(spec, repo)
	if err != nil {
		return nil, err
	}

	return &Report{
		SpecPath: specPath,
		RepoRoot: repoRoot,
		Endpoint: endpointReport,
		Model:    modelReport,
	}, nil
}

func buildEndpointReport(spec *openapi.Spec, repo *repoanalysis.Analyzer) (EndpointReport, error) {
	specOperations, err := spec.Operations()
	if err != nil {
		return EndpointReport{}, err
	}

	repoOperations := repo.Operations()
	specByKey := make(map[string]openapi.Operation, len(specOperations))
	for _, operation := range specOperations {
		specByKey[operationKey(operation.Method, operation.NormalizedPath)] = operation
	}

	repoByKey := make(map[string]repoanalysis.Operation, len(repoOperations))
	for _, operation := range repoOperations {
		key := operationKey(operation.Method, operation.NormalizedPath)
		if _, exists := repoByKey[key]; !exists {
			repoByKey[key] = operation
		}
	}

	report := EndpointReport{
		TotalSpecOperations:        len(specByKey),
		TotalImplementedOperations: len(repoByKey),
	}

	tagSummary := make(map[string]*TagCoverage)
	for _, operation := range specOperations {
		key := operationKey(operation.Method, operation.NormalizedPath)
		if _, ok := repoByKey[key]; ok {
			report.CoveredOperations++
		} else {
			report.Missing = append(report.Missing, operation)
		}

		tags := operation.Tags
		if len(tags) == 0 {
			tags = []string{"Uncategorized"}
		}

		for _, tag := range tags {
			coverage := tagSummary[tag]
			if coverage == nil {
				coverage = &TagCoverage{Tag: tag}
				tagSummary[tag] = coverage
			}

			coverage.Total++
			if _, ok := repoByKey[key]; ok {
				coverage.Covered++
			} else {
				coverage.Missing++
			}
		}
	}

	for _, operation := range repoByKey {
		if _, ok := specByKey[operationKey(operation.Method, operation.NormalizedPath)]; !ok {
			report.Extra = append(report.Extra, operation)
		}
	}

	for _, coverage := range tagSummary {
		report.ByTag = append(report.ByTag, *coverage)
	}

	slices.SortFunc(report.Missing, func(lhs, rhs openapi.Operation) int {
		if strings.Join(lhs.Tags, ",") != strings.Join(rhs.Tags, ",") {
			return strings.Compare(strings.Join(lhs.Tags, ","), strings.Join(rhs.Tags, ","))
		}
		if lhs.Path != rhs.Path {
			return strings.Compare(lhs.Path, rhs.Path)
		}
		return strings.Compare(lhs.Method, rhs.Method)
	})
	slices.SortFunc(report.Extra, func(lhs, rhs repoanalysis.Operation) int {
		if lhs.Path != rhs.Path {
			return strings.Compare(lhs.Path, rhs.Path)
		}
		return strings.Compare(lhs.Method, rhs.Method)
	})
	slices.SortFunc(report.ByTag, func(lhs, rhs TagCoverage) int {
		return strings.Compare(lhs.Tag, rhs.Tag)
	})

	return report, nil
}

func buildModelReport(spec *openapi.Spec, repo *repoanalysis.Analyzer) (ModelReport, error) {
	specModels, err := spec.Models()
	if err != nil {
		return ModelReport{}, err
	}

	repoModels, err := repo.Models()
	if err != nil {
		return ModelReport{}, err
	}

	repoIndex := make(map[string]repoanalysis.Model)
	duplicateRepoModels := make(map[string]bool)
	for _, model := range repoModels {
		key := normalizeModelName(model.Name)
		if _, exists := repoIndex[key]; exists {
			duplicateRepoModels[key] = true
			continue
		}
		repoIndex[key] = model
	}

	for key := range duplicateRepoModels {
		delete(repoIndex, key)
	}

	report := ModelReport{
		TotalSpecModels: len(specModels),
	}

	matchedRepoModels := make(map[string]bool)
	exactMatches := make(map[string]repoanalysis.Model)
	for _, specModel := range specModels {
		report.TotalSpecProperties += len(specModel.Properties)

		if repoModel, ok := repoIndex[normalizeModelName(specModel.Name)]; ok {
			exactMatches[specModel.Name] = repoModel
			matchedRepoModels[repoModel.Name] = true
		}
	}

	for _, specModel := range specModels {
		if repoModel, ok := exactMatches[specModel.Name]; ok {
			match := buildModelMatch(specModel, repoModel, "exact")
			report.CoveredModels++
			report.CoveredProperties += match.CoveredProperties
			report.Matches = append(report.Matches, match)
			if len(match.MissingProperties) > 0 || len(match.ExtraProperties) > 0 {
				report.ModelsWithGaps = append(report.ModelsWithGaps, match)
			}
			continue
		}

		repoModel, ok := heuristicModelMatch(specModel, repoModels, matchedRepoModels)
		if !ok {
			report.MissingModels = append(report.MissingModels, specModel)
			continue
		}

		matchedRepoModels[repoModel.Name] = true
		match := buildModelMatch(specModel, repoModel, "heuristic")
		report.CoveredModels++
		report.CoveredProperties += match.CoveredProperties
		report.Matches = append(report.Matches, match)
		if len(match.MissingProperties) > 0 || len(match.ExtraProperties) > 0 {
			report.ModelsWithGaps = append(report.ModelsWithGaps, match)
		}
	}

	for _, model := range repoModels {
		if matchedRepoModels[model.Name] {
			continue
		}
		report.ExtraModels = append(report.ExtraModels, model)
	}

	slices.SortFunc(report.Matches, func(lhs, rhs ModelMatch) int {
		return strings.Compare(lhs.Spec.Name, rhs.Spec.Name)
	})
	slices.SortFunc(report.ModelsWithGaps, func(lhs, rhs ModelMatch) int {
		return strings.Compare(lhs.Spec.Name, rhs.Spec.Name)
	})
	slices.SortFunc(report.MissingModels, func(lhs, rhs openapi.Model) int {
		return strings.Compare(lhs.Name, rhs.Name)
	})
	slices.SortFunc(report.ExtraModels, func(lhs, rhs repoanalysis.Model) int {
		return strings.Compare(lhs.Name, rhs.Name)
	})

	return report, nil
}

func buildModelMatch(specModel openapi.Model, repoModel repoanalysis.Model, matchType string) ModelMatch {
	repoFields := make(map[string]repoanalysis.Field, len(repoModel.Fields))
	for _, field := range repoModel.Fields {
		repoFields[field.Path] = field
	}

	match := ModelMatch{
		Spec:      specModel,
		Repo:      repoModel,
		MatchType: matchType,
	}

	for _, property := range specModel.Properties {
		if _, ok := repoFields[property]; ok {
			match.CoveredProperties++
			continue
		}
		match.MissingProperties = append(match.MissingProperties, property)
	}

	specFields := make(map[string]struct{}, len(specModel.Properties))
	for _, property := range specModel.Properties {
		specFields[property] = struct{}{}
	}

	for _, field := range repoModel.Fields {
		if _, ok := specFields[field.Path]; !ok {
			match.ExtraProperties = append(match.ExtraProperties, field)
		}
	}

	slices.Sort(match.MissingProperties)
	slices.SortFunc(match.ExtraProperties, func(lhs, rhs repoanalysis.Field) int {
		if lhs.Path != rhs.Path {
			return strings.Compare(lhs.Path, rhs.Path)
		}
		return lhs.Line - rhs.Line
	})

	return match
}

func WriteMarkdown(report *Report, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	files := map[string]string{
		"summary.md":                 renderSummary(report),
		"endpoint-coverage.md":       renderEndpointCoverage(report),
		"model-coverage.md":          renderModelCoverage(report),
		"model-property-coverage.md": renderModelPropertyCoverage(report),
	}

	for _, stale := range []string{"device-property-coverage.md"} {
		if err := os.Remove(filepath.Join(outputDir, stale)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale %s: %w", stale, err)
		}
	}

	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(outputDir, name), []byte(contents), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}

	return nil
}

func renderSummary(report *Report) string {
	var builder strings.Builder
	builder.WriteString("# Coverage Gap Summary\n\n")
	builder.WriteString("| Area | OpenAPI | Implemented | Missing |\n")
	builder.WriteString("| --- | ---: | ---: | ---: |\n")
	builder.WriteString(fmt.Sprintf("| Endpoint operations | %d | %d | %d |\n",
		report.Endpoint.TotalSpecOperations,
		report.Endpoint.CoveredOperations,
		len(report.Endpoint.Missing),
	))
	builder.WriteString(fmt.Sprintf("| API models | %d | %d | %d |\n",
		report.Model.TotalSpecModels,
		report.Model.CoveredModels,
		len(report.Model.MissingModels),
	))
	builder.WriteString(fmt.Sprintf("| API model properties | %d | %d | %d |\n",
		report.Model.TotalSpecProperties,
		report.Model.CoveredProperties,
		report.Model.TotalSpecProperties-report.Model.CoveredProperties,
	))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("- Spec: `%s`\n", report.SpecPath))
	builder.WriteString(fmt.Sprintf("- Repo: `%s`\n", report.RepoRoot))
	builder.WriteString("- Reports:\n")
	builder.WriteString("  - `endpoint-coverage.md`\n")
	builder.WriteString("  - `model-coverage.md`\n")
	builder.WriteString("  - `model-property-coverage.md`\n")
	return builder.String()
}

func renderEndpointCoverage(report *Report) string {
	var builder strings.Builder
	builder.WriteString("# Endpoint Coverage\n\n")
	builder.WriteString("| Tag | OpenAPI | Covered | Missing |\n")
	builder.WriteString("| --- | ---: | ---: | ---: |\n")
	for _, coverage := range report.Endpoint.ByTag {
		builder.WriteString(fmt.Sprintf("| %s | %d | %d | %d |\n", coverage.Tag, coverage.Total, coverage.Covered, coverage.Missing))
	}

	builder.WriteString("\n## Missing From The Client\n\n")
	if len(report.Endpoint.Missing) == 0 {
		builder.WriteString("No missing OpenAPI operations were found.\n")
	} else {
		builder.WriteString("| Method | Path | Operation ID | Tags | Summary |\n")
		builder.WriteString("| --- | --- | --- | --- | --- |\n")
		for _, operation := range report.Endpoint.Missing {
			builder.WriteString(fmt.Sprintf("| %s | `%s` | `%s` | %s | %s |\n",
				operation.Method,
				operation.Path,
				operation.OperationID,
				strings.Join(operation.Tags, ", "),
				operation.Summary,
			))
		}
	}

	builder.WriteString("\n## Implemented But Missing From OpenAPI\n\n")
	if len(report.Endpoint.Extra) == 0 {
		builder.WriteString("All implemented client operations matched an OpenAPI path.\n")
	} else {
		builder.WriteString("| Method | Path | Client Method | Source |\n")
		builder.WriteString("| --- | --- | --- | --- |\n")
		for _, operation := range report.Endpoint.Extra {
			builder.WriteString(fmt.Sprintf("| %s | `%s` | `%s` | `%s:%d` |\n",
				operation.Method,
				operation.Path,
				operation.ClientMethod,
				repoanalysis.RelativePath(report.RepoRoot, operation.File),
				operation.Line,
			))
		}
	}

	return builder.String()
}

func renderModelCoverage(report *Report) string {
	var builder strings.Builder
	builder.WriteString("# Model Coverage\n\n")
	builder.WriteString(fmt.Sprintf("- OpenAPI models: `%d`\n", report.Model.TotalSpecModels))
	builder.WriteString(fmt.Sprintf("- Matched client models: `%d`\n", report.Model.CoveredModels))
	builder.WriteString(fmt.Sprintf("- Missing client models: `%d`\n", len(report.Model.MissingModels)))
	builder.WriteString(fmt.Sprintf("- Extra client models: `%d`\n\n", len(report.Model.ExtraModels)))

	builder.WriteString("## Matched Models\n\n")
	if len(report.Model.Matches) == 0 {
		builder.WriteString("No OpenAPI models matched client structs.\n")
	} else {
		builder.WriteString("| OpenAPI Model | Client Model | Status | OpenAPI Properties | Covered | Missing | Extra |\n")
		builder.WriteString("| --- | --- | --- | ---: | ---: | ---: | ---: |\n")
		for _, match := range report.Model.Matches {
			status := "covered"
			if len(match.MissingProperties) > 0 || len(match.ExtraProperties) > 0 {
				status = "property gaps"
			}

			builder.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %d | %d | %d | %d |\n",
				match.Spec.Name,
				match.Repo.Name,
				statusWithMatchType(status, match.MatchType),
				len(match.Spec.Properties),
				match.CoveredProperties,
				len(match.MissingProperties),
				len(match.ExtraProperties),
			))
		}
	}

	builder.WriteString("\n## Missing OpenAPI Models\n\n")
	if len(report.Model.MissingModels) == 0 {
		builder.WriteString("No missing OpenAPI models were found.\n")
	} else {
		builder.WriteString("| OpenAPI Model | Properties |\n")
		builder.WriteString("| --- | ---: |\n")
		for _, model := range report.Model.MissingModels {
			builder.WriteString(fmt.Sprintf("| `%s` | %d |\n", model.Name, len(model.Properties)))
		}
	}

	builder.WriteString("\n## Extra Client Models\n\n")
	if len(report.Model.ExtraModels) == 0 {
		builder.WriteString("No extra client models were found.\n")
	} else {
		builder.WriteString("| Client Model | Source | Properties |\n")
		builder.WriteString("| --- | --- | ---: |\n")
		for _, model := range report.Model.ExtraModels {
			builder.WriteString(fmt.Sprintf("| `%s` | `%s:%d` | %d |\n",
				model.Name,
				repoanalysis.RelativePath(report.RepoRoot, model.File),
				model.Line,
				len(model.Fields),
			))
		}
	}

	return builder.String()
}

func renderModelPropertyCoverage(report *Report) string {
	var builder strings.Builder
	builder.WriteString("# Model Property Coverage\n\n")
	builder.WriteString(fmt.Sprintf("- OpenAPI model properties: `%d`\n", report.Model.TotalSpecProperties))
	builder.WriteString(fmt.Sprintf("- Covered in matched client models: `%d`\n", report.Model.CoveredProperties))
	builder.WriteString(fmt.Sprintf("- Missing from matched or absent client models: `%d`\n\n", report.Model.TotalSpecProperties-report.Model.CoveredProperties))

	builder.WriteString("## Missing OpenAPI Models\n\n")
	if len(report.Model.MissingModels) == 0 {
		builder.WriteString("No missing OpenAPI models were found.\n")
	} else {
		for _, model := range report.Model.MissingModels {
			builder.WriteString(fmt.Sprintf("### `%s`\n\n", model.Name))
			if len(model.Properties) == 0 {
				builder.WriteString("No leaf properties were extracted for this schema.\n\n")
				continue
			}

			builder.WriteString("| Property |\n")
			builder.WriteString("| --- |\n")
			for _, property := range model.Properties {
				builder.WriteString(fmt.Sprintf("| `%s` |\n", property))
			}
			builder.WriteString("\n")
		}
	}

	builder.WriteString("## Matched Models With Property Gaps\n\n")
	if len(report.Model.ModelsWithGaps) == 0 {
		builder.WriteString("No property gaps were found in matched models.\n")
	} else {
		for _, match := range report.Model.ModelsWithGaps {
			builder.WriteString(fmt.Sprintf("### `%s` -> `%s`\n\n", match.Spec.Name, match.Repo.Name))
			builder.WriteString(fmt.Sprintf("- Source: `%s:%d`\n", repoanalysis.RelativePath(report.RepoRoot, match.Repo.File), match.Repo.Line))
			builder.WriteString(fmt.Sprintf("- Match type: `%s`\n", match.MatchType))
			builder.WriteString(fmt.Sprintf("- OpenAPI properties: `%d`\n", len(match.Spec.Properties)))
			builder.WriteString(fmt.Sprintf("- Covered: `%d`\n", match.CoveredProperties))
			builder.WriteString(fmt.Sprintf("- Missing: `%d`\n", len(match.MissingProperties)))
			builder.WriteString(fmt.Sprintf("- Extra: `%d`\n\n", len(match.ExtraProperties)))

			if len(match.MissingProperties) > 0 {
				builder.WriteString("| Missing Property |\n")
				builder.WriteString("| --- |\n")
				for _, property := range match.MissingProperties {
					builder.WriteString(fmt.Sprintf("| `%s` |\n", property))
				}
				builder.WriteString("\n")
			}

			if len(match.ExtraProperties) > 0 {
				builder.WriteString("| Extra Property | Source |\n")
				builder.WriteString("| --- | --- |\n")
				for _, field := range match.ExtraProperties {
					builder.WriteString(fmt.Sprintf("| `%s` | `%s:%d` |\n",
						field.Path,
						repoanalysis.RelativePath(report.RepoRoot, field.File),
						field.Line,
					))
				}
				builder.WriteString("\n")
			}
		}
	}

	return builder.String()
}

func operationKey(method, path string) string {
	return method + " " + path
}

var nonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)

func normalizeModelName(name string) string {
	return nonAlphaNumeric.ReplaceAllString(strings.ToLower(name), "")
}

func heuristicModelMatch(specModel openapi.Model, repoModels []repoanalysis.Model, matched map[string]bool) (repoanalysis.Model, bool) {
	if len(specModel.Properties) == 0 {
		return repoanalysis.Model{}, false
	}

	bestScore := 0.0
	bestNameScore := 0.0
	var best repoanalysis.Model

	for _, candidate := range repoModels {
		if matched[candidate.Name] {
			continue
		}

		propertyScore := propertyCoverageScore(specModel.Properties, candidate.Fields)
		if propertyScore < 0.5 {
			continue
		}

		nameScore := modelNameTokenScore(specModel.Name, candidate.Name)
		if nameScore < 0.5 {
			continue
		}

		if propertyScore > bestScore || (propertyScore == bestScore && nameScore > bestNameScore) {
			best = candidate
			bestScore = propertyScore
			bestNameScore = nameScore
		}
	}

	if best.Name == "" {
		return repoanalysis.Model{}, false
	}

	return best, true
}

func propertyCoverageScore(specProperties []string, repoFields []repoanalysis.Field) float64 {
	if len(specProperties) == 0 {
		return 0
	}

	repoFieldSet := make(map[string]struct{}, len(repoFields))
	for _, field := range repoFields {
		repoFieldSet[field.Path] = struct{}{}
	}

	matches := 0
	for _, property := range specProperties {
		if _, ok := repoFieldSet[property]; ok {
			matches++
		}
	}

	return float64(matches) / float64(len(specProperties))
}

func modelNameTokenScore(specName, repoName string) float64 {
	specTokens := splitNameTokens(specName)
	repoTokens := make(map[string]struct{})
	for _, token := range splitNameTokens(repoName) {
		repoTokens[token] = struct{}{}
	}

	if len(specTokens) == 0 {
		return 0
	}

	matches := 0
	for _, token := range specTokens {
		if _, ok := repoTokens[token]; ok {
			matches++
		}
	}

	return float64(matches) / float64(len(specTokens))
}

func splitNameTokens(name string) []string {
	parts := regexp.MustCompile(`[A-Z]+[a-z0-9]*|[a-z0-9]+`).FindAllString(name, -1)
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		tokens = append(tokens, strings.ToLower(part))
	}
	return tokens
}

func statusWithMatchType(status, matchType string) string {
	if matchType == "exact" {
		return status
	}

	return status + " (" + matchType + ")"
}

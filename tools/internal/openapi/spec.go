package openapi

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Spec struct {
	document map[string]any
}

type Operation struct {
	Method         string
	Path           string
	NormalizedPath string
	OperationID    string
	Summary        string
	Tags           []string
}

type Model struct {
	Name       string
	Ref        string
	Properties []string
}

func Load(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec: %w", err)
	}

	return LoadBytes(data)
}

func LoadBytes(data []byte) (*Spec, error) {
	var document map[string]any
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}

	return &Spec{document: document}, nil
}

func NormalizePath(path string) string {
	path = strings.TrimPrefix(path, "/api/v2")
	if path == "" {
		path = "/"
	}

	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			segments[i] = "{}"
		}
	}

	return strings.Join(segments, "/")
}

func (s *Spec) Operations() ([]Operation, error) {
	paths, ok := mapValue(s.document["paths"])
	if !ok {
		return nil, fmt.Errorf("spec is missing paths")
	}

	pathKeys := sortedKeys(paths)
	operations := make([]Operation, 0)
	methods := []string{"get", "post", "put", "patch", "delete"}

	for _, path := range pathKeys {
		pathItem, ok := mapValue(paths[path])
		if !ok {
			continue
		}

		for _, method := range methods {
			operation, ok := mapValue(pathItem[method])
			if !ok {
				continue
			}

			operations = append(operations, Operation{
				Method:         strings.ToUpper(method),
				Path:           path,
				NormalizedPath: NormalizePath(path),
				OperationID:    stringValue(operation["operationId"]),
				Summary:        stringValue(operation["summary"]),
				Tags:           stringSlice(operation["tags"]),
			})
		}
	}

	return operations, nil
}

func (s *Spec) Models() ([]Model, error) {
	refs, err := s.reachableSchemaRefs()
	if err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(refs))
	for _, ref := range refs {
		schema, err := s.resolveRef(ref)
		if err != nil {
			return nil, err
		}

		if !isModelSchema(schema) {
			continue
		}

		properties := make(map[string]struct{})
		if err := s.collectLeafProperties(schema, "", properties, map[string]bool{}); err != nil {
			return nil, fmt.Errorf("collect properties for %s: %w", ref, err)
		}

		models = append(models, Model{
			Name:       schemaNameFromRef(ref),
			Ref:        ref,
			Properties: sortedSet(properties),
		})
	}

	slicesSortModels(models)
	return models, nil
}

func (s *Spec) operation(path, method string) (map[string]any, error) {
	paths, ok := mapValue(s.document["paths"])
	if !ok {
		return nil, fmt.Errorf("spec is missing paths")
	}

	pathItem, ok := mapValue(paths[path])
	if !ok {
		return nil, fmt.Errorf("path not found")
	}

	operation, ok := mapValue(pathItem[strings.ToLower(method)])
	if !ok {
		return nil, fmt.Errorf("operation not found")
	}

	return operation, nil
}

func (s *Spec) reachableSchemaRefs() ([]string, error) {
	paths, ok := mapValue(s.document["paths"])
	if !ok {
		return nil, fmt.Errorf("spec is missing paths")
	}

	refs := make(map[string]struct{})
	for _, path := range sortedKeys(paths) {
		pathItem, ok := mapValue(paths[path])
		if !ok {
			continue
		}

		for _, method := range []string{"get", "post", "put", "patch", "delete"} {
			operation, ok := mapValue(pathItem[method])
			if !ok {
				continue
			}

			if err := s.collectRefsFromOperation(operation, refs); err != nil {
				return nil, fmt.Errorf("%s %s: %w", strings.ToUpper(method), path, err)
			}
		}
	}

	return sortedSet(refs), nil
}

func (s *Spec) collectRefsFromOperation(operation map[string]any, refs map[string]struct{}) error {
	if requestBody, ok := operation["requestBody"]; ok {
		if err := s.collectRefsFromRequestBody(requestBody, refs); err != nil {
			return err
		}
	}

	responses, ok := mapValue(operation["responses"])
	if !ok {
		return nil
	}

	for _, status := range sortedKeys(responses) {
		if err := s.collectRefsFromResponse(responses[status], refs); err != nil {
			return err
		}
	}

	return nil
}

func (s *Spec) collectRefsFromRequestBody(requestBody any, refs map[string]struct{}) error {
	resolved, err := s.resolveMaybeRef(requestBody)
	if err != nil {
		return err
	}

	bodyMap, ok := mapValue(resolved)
	if !ok {
		return nil
	}

	content, ok := mapValue(bodyMap["content"])
	if !ok {
		return nil
	}

	for _, mediaType := range []string{"application/json", "application/hujson"} {
		if media, ok := mapValue(content[mediaType]); ok {
			if schema, ok := media["schema"]; ok {
				return s.collectSchemaRefs(schema, refs, map[string]bool{})
			}
		}
	}

	return nil
}

func (s *Spec) collectRefsFromResponse(response any, refs map[string]struct{}) error {
	resolved, err := s.resolveMaybeRef(response)
	if err != nil {
		return err
	}

	responseMap, ok := mapValue(resolved)
	if !ok {
		return nil
	}

	content, ok := mapValue(responseMap["content"])
	if !ok {
		return nil
	}

	if media, ok := mapValue(content["application/json"]); ok {
		if schema, ok := media["schema"]; ok {
			return s.collectSchemaRefs(schema, refs, map[string]bool{})
		}
	}

	return nil
}

func (s *Spec) collectSchemaRefs(schema any, refs map[string]struct{}, stack map[string]bool) error {
	schemaMap, ok := mapValue(schema)
	if !ok {
		return nil
	}

	if ref := stringValue(schemaMap["$ref"]); ref != "" {
		if stack[ref] {
			return nil
		}

		nextStack := copyStack(stack)
		nextStack[ref] = true

		if strings.HasPrefix(ref, "#/components/schemas/") {
			refs[ref] = struct{}{}
		}

		resolved, err := s.resolveRefAny(ref)
		if err != nil {
			return err
		}

		return s.collectSchemaRefs(resolved, refs, nextStack)
	}

	for _, combiner := range []string{"allOf", "anyOf", "oneOf"} {
		items, ok := sliceValue(schemaMap[combiner])
		if !ok {
			continue
		}

		for _, item := range items {
			if err := s.collectSchemaRefs(item, refs, stack); err != nil {
				return err
			}
		}
	}

	if properties, ok := mapValue(schemaMap["properties"]); ok {
		for _, name := range sortedKeys(properties) {
			if err := s.collectSchemaRefs(properties[name], refs, stack); err != nil {
				return err
			}
		}
	}

	if items, ok := schemaMap["items"]; ok {
		if err := s.collectSchemaRefs(items, refs, stack); err != nil {
			return err
		}
	}

	if additionalProperties, ok := schemaMap["additionalProperties"]; ok {
		if err := s.collectSchemaRefs(additionalProperties, refs, stack); err != nil {
			return err
		}
	}

	return nil
}

func (s *Spec) collectLeafProperties(schema any, prefix string, out map[string]struct{}, stack map[string]bool) error {
	schemaMap, ok := mapValue(schema)
	if !ok {
		if prefix != "" {
			out[prefix] = struct{}{}
		}
		return nil
	}

	if ref := stringValue(schemaMap["$ref"]); ref != "" {
		if stack[ref] {
			return nil
		}

		resolved, err := s.resolveRef(ref)
		if err != nil {
			return err
		}

		nextStack := copyStack(stack)
		nextStack[ref] = true
		return s.collectLeafProperties(resolved, prefix, out, nextStack)
	}

	structured := false
	for _, combiner := range []string{"allOf", "anyOf", "oneOf"} {
		items, ok := sliceValue(schemaMap[combiner])
		if !ok {
			continue
		}

		structured = true
		for _, item := range items {
			if err := s.collectLeafProperties(item, prefix, out, stack); err != nil {
				return err
			}
		}
	}

	if properties, ok := mapValue(schemaMap["properties"]); ok && len(properties) > 0 {
		structured = true
		for _, name := range sortedKeys(properties) {
			if err := s.collectLeafProperties(properties[name], joinPath(prefix, name), out, stack); err != nil {
				return err
			}
		}
	}

	if items, ok := mapValue(schemaMap["items"]); ok {
		structured = true
		if err := s.collectLeafProperties(items, prefix, out, stack); err != nil {
			return err
		}
	}

	if additionalProperties, ok := mapValue(schemaMap["additionalProperties"]); ok {
		structured = true
		if err := s.collectLeafProperties(additionalProperties, prefix, out, stack); err != nil {
			return err
		}
	}

	if !structured && prefix != "" {
		out[prefix] = struct{}{}
	}

	return nil
}

func (s *Spec) resolveRef(ref string) (map[string]any, error) {
	resolved, err := s.resolveRefAny(ref)
	if err != nil {
		return nil, err
	}

	mapped, ok := mapValue(resolved)
	if !ok {
		return nil, fmt.Errorf("ref %q does not resolve to an object", ref)
	}

	return mapped, nil
}

func (s *Spec) resolveRefAny(ref string) (any, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("unsupported ref %q", ref)
	}

	current := any(s.document)
	for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		node, ok := mapValue(current)
		if !ok {
			return nil, fmt.Errorf("ref %q does not resolve to an object", ref)
		}

		var exists bool
		current, exists = node[decodePointerToken(part)]
		if !exists {
			return nil, fmt.Errorf("ref %q is missing token %q", ref, part)
		}
	}

	return current, nil
}

func (s *Spec) resolveMaybeRef(value any) (any, error) {
	mapped, ok := mapValue(value)
	if !ok {
		return value, nil
	}

	if ref := stringValue(mapped["$ref"]); ref != "" {
		return s.resolveRefAny(ref)
	}

	return value, nil
}

func decodePointerToken(token string) string {
	token = strings.ReplaceAll(token, "~1", "/")
	token = strings.ReplaceAll(token, "~0", "~")
	return token
}

func mapValue(value any) (map[string]any, bool) {
	if value == nil {
		return nil, false
	}

	typed, ok := value.(map[string]any)
	return typed, ok
}

func sliceValue(value any) ([]any, bool) {
	if value == nil {
		return nil, false
	}

	typed, ok := value.([]any)
	return typed, ok
}

func stringValue(value any) string {
	typed, _ := value.(string)
	return typed
}

func stringSlice(value any) []string {
	values, ok := sliceValue(value)
	if !ok {
		return nil
	}

	out := make([]string, 0, len(values))
	for _, value := range values {
		if text := stringValue(value); text != "" {
			out = append(out, text)
		}
	}

	return out
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func sortedSet(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}

	return prefix + "." + name
}

func copyStack(values map[string]bool) map[string]bool {
	out := make(map[string]bool, len(values)+1)
	for key, value := range values {
		out[key] = value
	}

	return out
}

func schemaNameFromRef(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

func isModelSchema(schema map[string]any) bool {
	if _, ok := mapValue(schema["properties"]); ok {
		return true
	}
	if _, ok := schema["additionalProperties"]; ok {
		return true
	}
	if _, ok := schema["items"]; ok {
		return true
	}
	if schemaType := stringValue(schema["type"]); schemaType == "object" {
		return true
	}
	for _, combiner := range []string{"allOf", "anyOf", "oneOf"} {
		if _, ok := sliceValue(schema[combiner]); ok {
			return true
		}
	}

	return false
}

func slicesSortModels(models []Model) {
	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})
}

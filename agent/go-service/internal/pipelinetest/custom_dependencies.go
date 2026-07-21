package pipelinetest

import (
	"encoding/json"
	"fmt"
	"sort"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

func requiredCustomRecognitions(resource *maa.Resource, entry string) ([]string, error) {
	found := make(map[string]struct{})
	visited := make(map[string]struct{})
	var visitNode func(string) error
	var visitRecognition func(any) error

	visitNode = func(name string) error {
		if _, ok := visited[name]; ok {
			return nil
		}
		visited[name] = struct{}{}
		raw, err := resource.GetNodeJSON(name)
		if err != nil {
			return err
		}
		var node map[string]any
		if err := json.Unmarshal([]byte(raw), &node); err != nil {
			return fmt.Errorf("decode node %s: %w", name, err)
		}
		return visitRecognition(node["recognition"])
	}

	visitRecognition = func(value any) error {
		switch typed := value.(type) {
		case string:
			return nil
		case map[string]any:
			typeName, _ := typed["type"].(string)
			param, _ := typed["param"].(map[string]any)
			if typeName == "Custom" {
				if name, ok := param["custom_recognition"].(string); ok && name != "" {
					found[name] = struct{}{}
				}
			}
			for _, key := range []string{"all_of", "any_of"} {
				items, _ := param[key].([]any)
				for _, item := range items {
					if nodeName, ok := item.(string); ok {
						if err := visitNode(nodeName); err != nil {
							return err
						}
						continue
					}
					if err := visitRecognition(item); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	if err := visitNode(entry); err != nil {
		return nil, err
	}
	result := make([]string, 0, len(found))
	for name := range found {
		result = append(result, name)
	}
	sort.Strings(result)
	return result, nil
}

func missingRegistrations(resource *maa.Resource, required []string) ([]string, error) {
	registered, err := resource.GetCustomRecognitionList()
	if err != nil {
		return nil, err
	}
	available := make(map[string]struct{}, len(registered))
	for _, name := range registered {
		available[name] = struct{}{}
	}
	var missing []string
	for _, name := range required {
		if _, ok := available[name]; !ok {
			missing = append(missing, name)
		}
	}
	return missing, nil
}

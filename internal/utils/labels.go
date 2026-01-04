package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func LabelsMatch(objectLabels map[string]string, selector *metav1.LabelSelector) bool {
	if selector == nil {
		return true
	}
	for k, v := range selector.MatchLabels {
		if objectLabels[k] != v {
			return false
		}
	}
	return true
}

func ParseLabelSelector(raw string) map[string]string {
	result := make(map[string]string)
	if raw == "" {
		return result
	}
	parts := splitSelector(raw)
	for _, part := range parts {
		kv := splitKV(part)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

func splitSelector(s string) []string {
	var parts []string
	current := ""
	for _, c := range s {
		if c == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func splitKV(s string) []string {
	for i, c := range s {
		if c == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

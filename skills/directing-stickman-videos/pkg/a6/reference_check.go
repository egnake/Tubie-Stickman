package a6

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	placeholderRegex = regexp.MustCompile(`(?i)(\[[^\]\n]*(?:style|cast|world|character|scene)[ _-]*lock[^\]\n]*\]|\{\{[^}\n]+\}\}|\$\{[^}\n]+\}|<\s*(?:style|cast|world|character|scene|insert|prompt)\b[^>]*>|\b(?:tbd|todo|same\s+as\s+(?:above|before)|same\s+character|use\s+(?:the\s+)?previous\s+prompt)\b)`)
	promptFields     = []string{"style", "cast", "world", "composition", "constraints"}
)

// ImageInfo reads PNG/JPEG dimensions and computes SHA-256 without external dependencies.
func ImageInfo(path string) (int, int, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, "", fmt.Errorf("Unreadable image: %s: %w", path, err)
	}

	hash := sha256.Sum256(data)
	digest := hex.EncodeToString(hash[:])

	// Check PNG
	if len(data) >= 33 && string(data[:8]) == "\x89PNG\r\n\x1a\n" {
		if string(data[8:16]) != "\x00\x00\x00\rIHDR" {
			return 0, 0, "", fmt.Errorf("Invalid PNG header: %s", path)
		}
		w := int(binary.BigEndian.Uint32(data[16:20]))
		h := int(binary.BigEndian.Uint32(data[20:24]))
		if w <= 0 || h <= 0 {
			return 0, 0, "", fmt.Errorf("Invalid image dimensions: %s", path)
		}
		return w, h, digest, nil
	}

	// Check JPEG
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		sof := map[byte]bool{
			0xC0: true, 0xC1: true, 0xC2: true, 0xC3: true,
			0xC5: true, 0xC6: true, 0xC7: true,
			0xC9: true, 0xCA: true, 0xCB: true,
			0xCD: true, 0xCE: true, 0xCF: true,
		}
		index := 2
		var w, h int
		var found bool
		for index < len(data) {
			if data[index] != 0xFF {
				return 0, 0, "", fmt.Errorf("Invalid JPEG marker: %s", path)
			}
			for index < len(data) && data[index] == 0xFF {
				index++
			}
			if index >= len(data) {
				return 0, 0, "", fmt.Errorf("Truncated JPEG: %s", path)
			}
			marker := data[index]
			index++
			if marker == 0xD9 || marker == 0xDA {
				break
			}
			if marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7) {
				continue
			}
			if index+2 > len(data) {
				return 0, 0, "", fmt.Errorf("Truncated JPEG segment: %s", path)
			}
			segLen := int(binary.BigEndian.Uint16(data[index : index+2]))
			if segLen < 2 || index+segLen > len(data) {
				return 0, 0, "", fmt.Errorf("Invalid JPEG length: %s", path)
			}
			if sof[marker] {
				if segLen < 8 {
					return 0, 0, "", fmt.Errorf("Invalid JPEG size: %s", path)
				}
				h = int(binary.BigEndian.Uint16(data[index+3 : index+5]))
				w = int(binary.BigEndian.Uint16(data[index+5 : index+7]))
				found = true
				break
			}
			index += segLen
		}
		if !found {
			return 0, 0, "", fmt.Errorf("No JPEG dimensions: %s", path)
		}
		if w <= 0 || h <= 0 {
			return 0, 0, "", fmt.Errorf("Invalid image dimensions: %s", path)
		}
		return w, h, digest, nil
	}

	return 0, 0, "", fmt.Errorf("Not a readable PNG/JPEG header: %s", path)
}

// SafeSkillPath safely resolves a relative path within root, rejecting absolute paths, traversals, and symlinks escaping root.
func SafeSkillPath(root, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" {
		return "", errors.New("Missing relative file path")
	}

	// Reject absolute paths, .. traversal, backslashes and colons (URI / Windows drive specifiers)
	if filepath.IsAbs(rel) || strings.HasPrefix(rel, "/") || strings.Contains(rel, "\\") || strings.Contains(rel, ":") {
		return "", fmt.Errorf("Unsafe relative path: %s", rel)
	}

	parts := strings.Split(rel, "/")
	for _, p := range parts {
		if p == ".." {
			return "", fmt.Errorf("Unsafe relative path: %s", rel)
		}
	}

	base, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("Cannot resolve root %s: %w", root, err)
	}
	if resolvedBase, err := filepath.EvalSymlinks(base); err == nil {
		base = resolvedBase
	}

	target := filepath.Join(base, filepath.FromSlash(rel))
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", fmt.Errorf("Missing/unsafe resource: %s", rel)
	}

	fi, err := os.Stat(resolvedTarget)
	if err != nil || fi.IsDir() {
		return "", fmt.Errorf("Missing/unsafe resource: %s", rel)
	}

	relToRoot, err := filepath.Rel(base, resolvedTarget)
	if err != nil || strings.HasPrefix(relToRoot, "..") || relToRoot == ".." || filepath.IsAbs(relToRoot) {
		return "", fmt.Errorf("Missing/unsafe resource: %s", rel)
	}

	return resolvedTarget, nil
}

func readJSON(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Cannot read JSON %s: %w", path, err)
	}
	var val any
	if err := json.Unmarshal(data, &val); err != nil {
		return nil, fmt.Errorf("Cannot read JSON %s: %w", path, err)
	}
	return val, nil
}

func fullText(value any, label string) error {
	str, ok := value.(string)
	if !ok || strings.TrimSpace(str) == "" {
		return fmt.Errorf("Empty prompt field: %s", label)
	}
	if match := placeholderRegex.FindString(str); match != "" {
		return fmt.Errorf("Unexpanded placeholder in %s: %s", label, match)
	}
	return nil
}

// ValidatePack validates reference-manifest.json and all bundled A6 assets within root.
func ValidatePack(root string) ([]map[string]any, error) {
	manifestPath, err := SafeSkillPath(root, "assets/a6/reference-manifest.json")
	if err != nil {
		return nil, err
	}
	raw, err := readJSON(manifestPath)
	if err != nil {
		return nil, err
	}
	manifest, ok := raw.(map[string]any)
	if !ok || manifest["schema"] != "a6-reference-pack-v1" {
		return nil, errors.New("Unsupported A6 manifest")
	}

	rawAssets, ok := manifest["assets"].([]any)
	if !ok || len(rawAssets) < 1 {
		return nil, errors.New("No A6 reference assets")
	}

	foundMap := make(map[string]map[string]any)
	seenPaths := make(map[string]bool)
	var assetList []map[string]any

	for _, item := range rawAssets {
		asset, ok := item.(map[string]any)
		if !ok {
			return nil, errors.New("Invalid reference asset")
		}
		key, ok := asset["id"].(string)
		if !ok || key == "" || foundMap[key] != nil {
			return nil, errors.New("Missing/duplicate asset ID")
		}
		path, ok := asset["path"].(string)
		if !ok || !strings.HasPrefix(path, "assets/a6/") {
			return nil, errors.New("A6 asset outside assets/a6")
		}
		if seenPaths[path] {
			return nil, errors.New("Duplicate A6 asset path")
		}
		localFile, err := SafeSkillPath(root, path)
		if err != nil {
			return nil, err
		}
		w, h, digest, err := ImageInfo(localFile)
		if err != nil {
			return nil, err
		}
		expectedSha, _ := asset["sha256"].(string)
		if digest != expectedSha {
			return nil, fmt.Errorf("SHA-256 mismatch: %s", path)
		}
		expectedW, _ := asset["width"].(float64)
		expectedH, _ := asset["height"].(float64)
		if w != int(expectedW) || h != int(expectedH) {
			return nil, fmt.Errorf("Dimension mismatch: %s", path)
		}

		uses, ok := asset["uses"].([]any)
		hasStyle := false
		if ok {
			for _, u := range uses {
				if uStr, ok := u.(string); ok && uStr == "style" {
					hasStyle = true
					break
				}
			}
		}
		if !hasStyle {
			return nil, fmt.Errorf("Missing style use: %s", key)
		}

		for _, field := range []string{"provenance", "acceptance"} {
			v, ok := asset[field].(string)
			if !ok || strings.TrimSpace(v) == "" {
				return nil, fmt.Errorf("Missing %s: %s", field, key)
			}
		}

		foundMap[key] = asset
		seenPaths[path] = true
		assetList = append(assetList, asset)
	}

	return assetList, nil
}

// ValidatePlan validates an A6 reference plan against bundled package references and current project assets.
func ValidatePlan(root, planPath string) ([]map[string]any, error) {
	return ValidatePlanWithProject(root, planPath, "")
}

// ValidatePlanWithProject validates an A6 plan with an explicit projectRoot.
func ValidatePlanWithProject(root, planPath, projectRoot string) ([]map[string]any, error) {
	assetsList, err := ValidatePack(root)
	if err != nil {
		return nil, err
	}
	packageAssets := make(map[string]map[string]any)
	for _, a := range assetsList {
		if id, ok := a["id"].(string); ok {
			packageAssets[id] = a
		}
	}

	raw, err := readJSON(planPath)
	if err != nil {
		return nil, err
	}
	plan, ok := raw.(map[string]any)
	if !ok || plan["schema"] != "a6-reference-plan-v1" {
		return nil, errors.New("Unsupported plan schema")
	}

	rawImages, ok := plan["images"].([]any)
	if !ok || len(rawImages) == 0 {
		return nil, errors.New("No planned images")
	}

	base := projectRoot
	if base == "" {
		resolvedPlan, err := filepath.Abs(planPath)
		if err == nil {
			base = filepath.Dir(resolvedPlan)
		} else {
			base = filepath.Dir(planPath)
		}
	}

	seenIDs := make(map[string]bool)
	var expanded []map[string]any

	for _, item := range rawImages {
		img, ok := item.(map[string]any)
		if !ok {
			return nil, errors.New("Invalid planned image")
		}
		key, ok := img["id"].(string)
		if !ok || strings.TrimSpace(key) == "" || seenIDs[key] {
			return nil, errors.New("Missing/duplicate image ID")
		}
		seenIDs[key] = true

		purpose, ok := img["purpose"].(string)
		validPurposes := map[string]bool{
			"character_reference":   true,
			"environment_reference": true,
			"shot_first_frame":      true,
		}
		if !ok || !validPurposes[purpose] {
			return nil, fmt.Errorf("Invalid image purpose: %s", key)
		}

		promptMap, ok := img["image_prompt"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("Missing complete image_prompt: %s", key)
		}
		var fieldValues []string
		for _, field := range promptFields {
			if err := fullText(promptMap[field], key+"."+field); err != nil {
				return nil, err
			}
			fieldValues = append(fieldValues, strings.TrimSpace(promptMap[field].(string)))
		}

		if motion, exists := img["motion_prompt"]; exists {
			if err := fullText(motion, key+".motion_prompt"); err != nil {
				return nil, err
			}
		}

		refs, ok := img["reference_inputs"].([]any)
		if !ok || len(refs) == 0 {
			return nil, fmt.Errorf("No reference inputs: %s", key)
		}

		for _, r := range refs {
			ref, ok := r.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("Missing/invalid reference input: %s", key)
			}
			role, _ := ref["role"].(string)
			if role != "style" && role != "identity" && role != "first_frame" {
				return nil, fmt.Errorf("Missing/invalid reference role: %s", key)
			}
			source, _ := ref["source"].(string)
			if source == "package" {
				assetID, _ := ref["asset_id"].(string)
				if packageAssets[assetID] == nil {
					return nil, fmt.Errorf("Unknown package reference: %s", key)
				}
				if role == "first_frame" {
					return nil, fmt.Errorf("Package example is not a new-story first frame: %s", key)
				}
				reuseSample, _ := plan["reuse_sample_identity"].(bool)
				if role == "identity" && !reuseSample {
					return nil, fmt.Errorf("Sample identity was not explicitly chosen: %s", key)
				}
			} else if source == "project" {
				p, _ := ref["path"].(string)
				localProjFile, err := SafeSkillPath(base, p)
				if err != nil {
					return nil, err
				}
				_, _, digest, err := ImageInfo(localProjFile)
				if err != nil {
					return nil, err
				}
				if expectedSha, exists := ref["sha256"].(string); exists {
					if digest != expectedSha {
						return nil, fmt.Errorf("Project reference hash mismatch: %s", key)
					}
				}
			} else {
				return nil, fmt.Errorf("Invalid reference source: %s", key)
			}
		}

		expandedText := strings.Join(fieldValues, "\n")
		expanded = append(expanded, map[string]any{
			"id":     key,
			"key":    key,
			"prompt": expandedText,
		})
	}

	return expanded, nil
}

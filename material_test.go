package obj

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

// TestMaterialParsingFromString tests material parsing without external files
func TestMaterialParsingFromString(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]*Material
		wantErr  bool
	}{
		{
			name: "Basic material",
			content: `# Basic material test
newmtl basic_material
Ka 0.2 0.2 0.2
Kd 0.8 0.8 0.8
Ks 0.0 0.0 0.0
Ns 32.0
d 1.0
illum 2
`,
			expected: map[string]*Material{
				"basic_material": {
					Name:         "basic_material",
					Ambient:      []float32{0.2, 0.2, 0.2, 1.0},
					Diffuse:      []float32{1.0, 1.0, 1.0, 1.0}, // Multiplied by 1.3
					Specular:     []float32{0.0, 0.0, 0.0, 1.0},
					Shininess:    0.032,
					Opacity:      1.0,
					Illumination: 2,
					Emissive:     []float32{0.2, 0.2, 0.2, 1.0},
				},
			},
			wantErr: false,
		},
		{
			name: "Material with textures",
			content: `# Material with textures
newmtl textured_material
Ka 0.1 0.1 0.1
Kd 0.6 0.4 0.2
Ks 0.9 0.9 0.9
Ns 100.0
map_Ka ambient.jpg
map_Kd diffuse.jpg
map_Ks specular.jpg
map_d alpha.png
map_bump normal.jpg
`,
			expected: map[string]*Material{
				"textured_material": {
					Name:            "textured_material",
					Ambient:         []float32{0.1, 0.1, 0.1, 1.0},
					Diffuse:         []float32{0.78, 0.52, 0.26, 1.0}, // Multiplied by 1.3
					Specular:        []float32{0.9, 0.9, 0.9, 1.0},
					Shininess:       0.1,
					Opacity:         1.0,
					AmbientTexture:  "ambient.jpg",
					DiffuseTexture:  "diffuse.jpg",
					SpecularTexture: "specular.jpg",
					AlphaTexture:    "alpha.png",
					BumpTexture:     "normal.jpg",
					Emissive:        []float32{0.2, 0.2, 0.2, 1.0},
				},
			},
			wantErr: false,
		},
		{
			name: "PBR material",
			content: `# PBR material
newmtl pbr_material
Ka 0.05 0.05 0.05
Kd 0.7 0.2 0.1
Ks 0.01 0.01 0.01
Pr 0.5
Pm 0.8
Ps 0.1
Pc 0.2
Pcr 0.1
aniso 0.5
anisor 0.3
`,
			expected: map[string]*Material{
				"pbr_material": {
					Name:               "pbr_material",
					Ambient:            []float32{0.05, 0.05, 0.05, 1.0},
					Diffuse:            []float32{0.91, 0.26, 0.13, 1.0}, // Multiplied by 1.3
					Specular:           []float32{0.01, 0.01, 0.01, 1.0},
					Shininess:          0.0,
					Opacity:            1.0,
					Roughness:          0.5,
					Metallic:           0.8,
					Sheen:              0.1,
					ClearcoatThickness: 0.2,
					ClearcoatRoughness: 0.1,
					Anisotropy:         0.5,
					AnisotropyRotation: 0.3,
					Emissive:           []float32{0.2, 0.2, 0.2, 1.0},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple materials",
			content: `# First material
newmtl material1
Ka 0.1 0.1 0.1
Kd 0.5 0.5 0.5

# Second material
newmtl material2
Ka 0.2 0.2 0.2
Kd 0.8 0.2 0.2
`,
			expected: map[string]*Material{
				"material1": {
					Name:     "material1",
					Ambient:  []float32{0.1, 0.1, 0.1, 1.0},
					Diffuse:  []float32{0.65, 0.65, 0.65, 1.0}, // Multiplied by 1.3
					Specular: []float32{0.0, 0.0, 0.0, 1.0},
					Opacity:  1.0,
					Emissive: []float32{0.2, 0.2, 0.2, 1.0},
				},
				"material2": {
					Name:     "material2",
					Ambient:  []float32{0.2, 0.2, 0.2, 1.0},
					Diffuse:  []float32{1.0, 0.26, 0.26, 1.0}, // Multiplied by 1.3
					Specular: []float32{0.0, 0.0, 0.0, 1.0},
					Opacity:  1.0,
					Emissive: []float32{0.2, 0.2, 0.2, 1.0},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid format",
			content: `# Invalid format
newmtl invalid
Ka 0.1 0.1
`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader from string instead of file
			reader := strings.NewReader(tt.content)

			// Parse materials from reader
			materials, err := readMaterialsFromReader(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("readMaterialsFromReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(materials) != len(tt.expected) {
					t.Errorf("Expected %d materials, got %d", len(tt.expected), len(materials))
					return
				}

				for name, expectedMat := range tt.expected {
					gotMat, exists := materials[name]
					if !exists {
						t.Errorf("Material %s not found", name)
						continue
					}

					if gotMat.Name != expectedMat.Name {
						t.Errorf("Material name mismatch: expected %s, got %s", expectedMat.Name, gotMat.Name)
					}

					// Compare float slices with tolerance
					compareFloatSlices(t, "Ambient", gotMat.Ambient, expectedMat.Ambient)
					compareFloatSlices(t, "Diffuse", gotMat.Diffuse, expectedMat.Diffuse)
					compareFloatSlices(t, "Specular", gotMat.Specular, expectedMat.Specular)

					// Compare other fields
					if !floatEqual(gotMat.Shininess, expectedMat.Shininess) {
						t.Errorf("Shininess mismatch: expected %f, got %f", expectedMat.Shininess, gotMat.Shininess)
					}
					if !floatEqual(gotMat.Opacity, expectedMat.Opacity) {
						t.Errorf("Opacity mismatch: expected %f, got %f", expectedMat.Opacity, gotMat.Opacity)
					}
					if gotMat.AmbientTexture != expectedMat.AmbientTexture {
						t.Errorf("AmbientTexture mismatch: expected %s, got %s", expectedMat.AmbientTexture, gotMat.AmbientTexture)
					}
					if gotMat.DiffuseTexture != expectedMat.DiffuseTexture {
						t.Errorf("DiffuseTexture mismatch: expected %s, got %s", expectedMat.DiffuseTexture, gotMat.DiffuseTexture)
					}
					if gotMat.SpecularTexture != expectedMat.SpecularTexture {
						t.Errorf("SpecularTexture mismatch: expected %s, got %s", expectedMat.SpecularTexture, gotMat.SpecularTexture)
					}
					if gotMetallic := gotMat.Metallic; !floatEqual(float64(gotMetallic), float64(expectedMat.Metallic)) {
						t.Errorf("Metallic mismatch: expected %f, got %f", expectedMat.Metallic, gotMetallic)
					}
					if gotRoughness := gotMat.Roughness; !floatEqual(float64(gotRoughness), float64(expectedMat.Roughness)) {
						t.Errorf("Roughness mismatch: expected %f, got %f", expectedMat.Roughness, gotRoughness)
					}
				}
			}
		})
	}
}

// readMaterialsFromReader reads materials from an io.Reader instead of a file
func readMaterialsFromReader(reader *strings.Reader) (map[string]*Material, error) {
	var (
		materials = make(map[string]*Material)
		material  *Material
	)

	lno := 0
	line := ""
	scanner := bufio.NewScanner(reader)

	fail := func(msg string) error {
		return fmt.Errorf(msg+" at line %d: %s", lno, line)
	}

	for scanner.Scan() {
		lno++
		line = scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if fields[0] == "newmtl" {
			if len(fields) != 2 {
				return nil, fail("unsupported material definition")
			}

			material = &Material{Name: fields[1]}
			material.Ambient = []float32{0.0, 0.0, 0.0, 1.0}
			material.Diffuse = []float32{0.8, 0.8, 0.8, 1.0}
			material.Specular = []float32{0.0, 0.0, 0.0, 1.0}
			material.TransmissionFilter = []float32{1.0, 1.0, 1.0}
			material.Emissive = []float32{0.2, 0.2, 0.2, 1.0}
			material.Opacity = 1
			materials[material.Name] = material

			continue
		}

		if material == nil {
			return nil, fail("found data before material")
		}

		// Parse material properties (same as ReadMaterials)
		switch fields[0] {
		case "Ka":
			if len(fields) != 4 {
				return nil, fail("unsupported ambient color line")
			}
			for i := 0; i < 3; i++ {
				f, err := strconv.ParseFloat(fields[i+1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.Ambient[i] = float32(f)
			}
		case "Kd":
			if len(fields) != 4 {
				return nil, fail("unsupported diffuse color line")
			}
			for i := 0; i < 3; i++ {
				f, err := strconv.ParseFloat(fields[i+1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.Diffuse[i] = float32(f)
			}
		case "Ks":
			if len(fields) != 4 {
				return nil, fail("unsupported specular color line")
			}
			for i := 0; i < 3; i++ {
				f, err := strconv.ParseFloat(fields[i+1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.Specular[i] = float32(f)
			}
		case "Ke":
			if len(fields) != 4 {
				return nil, fail("unsupported emissive color line")
			}
			for i := 0; i < 3; i++ {
				f, err := strconv.ParseFloat(fields[i+1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				if f != 0 {
					material.Emissive[i] = float32(f)
				}
			}
		case "Ns":
			if len(fields) != 2 {
				return nil, fail("unsupported shininess line")
			}
			f, err := strconv.ParseFloat(fields[1], 32)
			if err != nil {
				return nil, fail("cannot parse float")
			}
			material.Shininess = float64(f / 1000)
		case "d":
			if len(fields) != 2 {
				return nil, fail("unsupported transparency line")
			}
			f, err := strconv.ParseFloat(fields[1], 32)
			if err != nil {
				return nil, fail("cannot parse float")
			}
			material.Opacity = f
		case "map_Ka":
			if len(fields) == 2 {
				material.AmbientTexture = fields[1]
			}
		case "map_Kd":
			if len(fields) == 2 {
				material.DiffuseTexture = fields[1]
			}
		case "map_Ks":
			if len(fields) == 2 {
				material.SpecularTexture = fields[1]
			}
		case "map_Ns":
		case "map_Ke":
			if len(fields) == 2 {
				material.EmissiveTexture = fields[1]
			}
		case "map_d":
		case "map_opacity":
			if len(fields) == 2 {
				material.AlphaTexture = fields[1]
			}
		case "map_bump":
		case "bump":
			if len(fields) == 2 {
				material.BumpTexture = fields[1]
			}
		case "illum":
			if len(fields) == 2 {
				f, err := strconv.ParseUint(fields[1], 0, 10)
				if err != nil {
					return nil, fail("cannot parse uint")
				}
				material.Illumination = uint32(f)
			}
		case "Pr":
			if len(fields) == 2 {
				f, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.Roughness = float32(f)
			}
		case "Pm":
			if len(fields) == 2 {
				f, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.Metallic = float32(f)
			}
		case "Ps":
			if len(fields) == 2 {
				f, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.Sheen = float32(f)
			}
		case "Pc":
			if len(fields) == 2 {
				f, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.ClearcoatThickness = float32(f)
			}
		case "Pcr":
			if len(fields) == 2 {
				f, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.ClearcoatRoughness = float32(f)
			}
		case "aniso":
			if len(fields) == 2 {
				f, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.Anisotropy = float32(f)
			}
		case "anisor":
			if len(fields) == 2 {
				f, err := strconv.ParseFloat(fields[1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.AnisotropyRotation = float32(f)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Apply the same post-processing as ReadMaterials
	for _, material := range materials {
		for i := 0; i < 3; i++ {
			material.Diffuse[i] *= 1.3
			if material.Diffuse[i] > 1 {
				material.Diffuse[i] = 1
			}
		}
	}

	return materials, nil
}

// TestMaterialWriteRead tests round-trip write/read functionality
func TestMaterialWriteRead(t *testing.T) {
	originalMaterials := map[string]*Material{
		"test_material": {
			Name:               "test_material",
			Ambient:            []float32{0.1, 0.2, 0.3, 1.0},
			Diffuse:            []float32{0.4, 0.5, 0.6, 1.0},
			Specular:           []float32{0.7, 0.8, 0.9, 1.0},
			Emissive:           []float32{0.2, 0.2, 0.2, 1.0},
			TransmissionFilter: []float32{1.0, 1.0, 1.0},
			Shininess:          0.05,
			Opacity:            0.9,
			Illumination:       2,
			Roughness:          0.3,
			Metallic:           0.7,
			Sheen:              0.1,
			ClearcoatThickness: 0.2,
			ClearcoatRoughness: 0.05,
			Anisotropy:         0.4,
			AnisotropyRotation: 0.2,
			AmbientTexture:     "test_ambient.jpg",
			DiffuseTexture:     "test_diffuse.jpg",
			SpecularTexture:    "test_specular.jpg",
			EmissiveTexture:    "test_emissive.jpg",
			AlphaTexture:       "test_alpha.png",
			BumpTexture:        "test_bump.jpg",
		},
	}

	// Write materials to buffer
	var buf bytes.Buffer
	writer := &materialWriter{buffer: &buf}

	for name, mat := range originalMaterials {
		writer.writeMaterial(name, mat)
	}

	// Read materials back from buffer
	reader := strings.NewReader(buf.String())
	readMaterials, err := readMaterialsFromReader(reader)
	if err != nil {
		t.Fatalf("Failed to read materials: %v", err)
	}

	// Compare results
	if len(readMaterials) != len(originalMaterials) {
		t.Errorf("Expected %d materials, got %d", len(originalMaterials), len(readMaterials))
	}

	for name, expected := range originalMaterials {
		got, exists := readMaterials[name]
		if !exists {
			t.Errorf("Material %s not found after round-trip", name)
			continue
		}

		// Note: Due to the 1.3 multiplication in ReadMaterials, we need to adjust expectations
		adjustedDiffuse := make([]float32, len(expected.Diffuse))
		copy(adjustedDiffuse, expected.Diffuse)
		for i := 0; i < 3; i++ {
			adjustedDiffuse[i] *= 1.3
			if adjustedDiffuse[i] > 1 {
				adjustedDiffuse[i] = 1
			}
		}

		compareFloatSlices(t, "Diffuse", got.Diffuse, adjustedDiffuse)
		compareFloatSlices(t, "Ambient", got.Ambient, expected.Ambient)
		compareFloatSlices(t, "Specular", got.Specular, expected.Specular)
		compareFloatSlices(t, "Emissive", got.Emissive, expected.Emissive)
	}
}

// TestMaterialDefaults tests default values
func TestMaterialDefaults(t *testing.T) {
	content := `newmtl default_test
`

	reader := strings.NewReader(content)
	materials, err := readMaterialsFromReader(reader)
	if err != nil {
		t.Fatalf("Failed to parse material: %v", err)
	}

	if len(materials) != 1 {
		t.Fatalf("Expected 1 material, got %d", len(materials))
	}

	mat := materials["default_test"]

	// Check defaults
	expectedAmbient := []float32{0.0, 0.0, 0.0, 1.0}
	expectedDiffuse := []float32{1.0, 1.0, 1.0, 1.0} // After multiplication
	expectedSpecular := []float32{0.0, 0.0, 0.0, 1.0}

	compareFloatSlices(t, "Ambient", mat.Ambient, expectedAmbient)
	compareFloatSlices(t, "Diffuse", mat.Diffuse, expectedDiffuse)
	compareFloatSlices(t, "Specular", mat.Specular, expectedSpecular)

	if mat.Opacity != 1.0 {
		t.Errorf("Expected opacity 1.0, got %f", mat.Opacity)
	}
	if mat.Shininess != 0.0 {
		t.Errorf("Expected shininess 0.0, got %f", mat.Shininess)
	}
	if mat.Illumination != 0 {
		t.Errorf("Expected illumination 0, got %d", mat.Illumination)
	}
}

// Helper functions

func compareFloatSlices(t *testing.T, name string, got, expected []float32) {
	if len(got) != len(expected) {
		t.Errorf("%s length mismatch: expected %d, got %d", name, len(expected), len(got))
		return
	}

	for i := 0; i < len(expected); i++ {
		if !floatEqual(float64(got[i]), float64(expected[i])) {
			t.Errorf("%s[%d] mismatch: expected %f, got %f", name, i, expected[i], got[i])
		}
	}
}

func floatEqual(a, b float64) bool {
	const epsilon = 1e-6
	return abs(a-b) < epsilon
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// materialWriter helps write material content for testing

type materialWriter struct {
	buffer *bytes.Buffer
}

func (w *materialWriter) writeString(s string) {
	w.buffer.WriteString(s)
}

func (w *materialWriter) writeMaterial(name string, mat *Material) {
	w.writeString(fmt.Sprintf("newmtl %s\n", name))

	if len(mat.Ambient) >= 3 {
		w.writeString(fmt.Sprintf("Ka %g %g %g\n", mat.Ambient[0], mat.Ambient[1], mat.Ambient[2]))
	}
	if len(mat.Diffuse) >= 3 {
		w.writeString(fmt.Sprintf("Kd %g %g %g\n", mat.Diffuse[0], mat.Diffuse[1], mat.Diffuse[2]))
	}
	if len(mat.Specular) >= 3 {
		w.writeString(fmt.Sprintf("Ks %g %g %g\n", mat.Specular[0], mat.Specular[1], mat.Specular[2]))
	}
	if !math.IsNaN(mat.Shininess) {
		w.writeString(fmt.Sprintf("Ns %g\n", mat.Shininess*1000))
	}
	if !math.IsNaN(mat.Opacity) {
		w.writeString(fmt.Sprintf("d %g\n", mat.Opacity))
	}
	if mat.AmbientTexture != "" {
		w.writeString(fmt.Sprintf("map_Ka %s\n", mat.AmbientTexture))
	}
	if mat.DiffuseTexture != "" {
		w.writeString(fmt.Sprintf("map_Kd %s\n", mat.DiffuseTexture))
	}
	if mat.SpecularTexture != "" {
		w.writeString(fmt.Sprintf("map_Ks %s\n", mat.SpecularTexture))
	}
	if mat.AlphaTexture != "" {
		w.writeString(fmt.Sprintf("map_d %s\n", mat.AlphaTexture))
	}
	if mat.BumpTexture != "" {
		w.writeString(fmt.Sprintf("map_bump %s\n", mat.BumpTexture))
	}
	if mat.Roughness != 0 {
		w.writeString(fmt.Sprintf("Pr %g\n", mat.Roughness))
	}
	if mat.Metallic != 0 {
		w.writeString(fmt.Sprintf("Pm %g\n", mat.Metallic))
	}
	if mat.Sheen != 0 {
		w.writeString(fmt.Sprintf("Ps %g\n", mat.Sheen))
	}
	if mat.ClearcoatThickness != 0 {
		w.writeString(fmt.Sprintf("Pc %g\n", mat.ClearcoatThickness))
	}
	if mat.ClearcoatRoughness != 0 {
		w.writeString(fmt.Sprintf("Pcr %g\n", mat.ClearcoatRoughness))
	}
	if mat.Anisotropy != 0 {
		w.writeString(fmt.Sprintf("aniso %g\n", mat.Anisotropy))
	}
	if mat.AnisotropyRotation != 0 {
		w.writeString(fmt.Sprintf("anisor %g\n", mat.AnisotropyRotation))
	}
	w.writeString("\n")
}

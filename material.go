package obj

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type Material struct {
	Name               string
	Ambient            []float32
	Diffuse            []float32
	Specular           []float32
	Emissive           []float32
	TransmissionFilter []float32
	Shininess          float64
	AmbientTexture     string
	DiffuseTexture     string
	SpecularTexture    string
	EmissiveTexture    string
	AlphaTexture       string
	BumpTexture        string
	Opacity            float64
	Illumination       uint32
	Roughness          float32
	Metallic           float32
	Sheen              float32
	ClearcoatThickness float32
	ClearcoatRoughness float32
	Anisotropy         float32
	AnisotropyRotation float32
}

func ReadMaterials(filename string) (map[string]*Material, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot read referenced material library: %v", err)
	}
	defer file.Close()

	var (
		materials = make(map[string]*Material)
		material  *Material
	)

	lno := 0
	line := ""
	scanner := bufio.NewScanner(file)

	fail := func(msg string) error {
		return fmt.Errorf(msg+" at %s:%d: %s", filename, lno, line)
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
				return nil, fail("unsupported specular color line")
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
		case "Tf":
			if len(fields) != 4 {
				return nil, fail("unsupported transmission filter line")
			}
			for i := 0; i < 3; i++ {
				f, err := strconv.ParseFloat(fields[i+1], 32)
				if err != nil {
					return nil, fail("cannot parse float")
				}
				material.TransmissionFilter[i] = float32(f)
			}
		case "map_Ka":
			if len(fields) == 2 {
				material.AmbientTexture = fields[1]
			}
		case "map_Kd":
			if len(fields) == 2 {
				material.DiffuseTexture = fields[1]
			}
		case "map_Ns":
		case "map_Ks":
			if len(fields) == 2 {
				material.SpecularTexture = fields[1]
			}
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
		case "refl":
			if len(fields) == 2 {
				f, err := strconv.ParseUint(fields[1], 0, 10)
				if err != nil {
					return nil, fail("cannot parse float")
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

func WriteMaterials(filename string, mtls map[string]*Material) error {
	var ret []byte
	buff := bytes.NewBuffer(ret)
	_, err := buff.WriteString("#\n")
	if err != nil {
		return err
	}
	_, err = buff.WriteString("# Wavefront material file\n")
	if err != nil {
		return err
	}
	_, err = buff.WriteString("# Model Gen for flywave\n")
	if err != nil {
		return err
	}
	_, err = buff.WriteString("#\n")
	if err != nil {
		return err
	}

	for i, k := range mtls {
		_, err = buff.WriteString("\n")
		if err != nil {
			return err
		}
		buff.WriteString(fmt.Sprintf("newmtl %s\n", i))
		if k.Ambient != nil {
			_, err = buff.WriteString(fmt.Sprintf("Ka %g %g %g\n", k.Ambient[0], k.Ambient[1], k.Ambient[2]))
			if err != nil {
				return err
			}
		}
		if k.Diffuse != nil {
			_, err = buff.WriteString(fmt.Sprintf("Kd %g %g %g\n", k.Diffuse[0], k.Diffuse[1], k.Diffuse[2]))
			if err != nil {
				return err
			}
		}
		if k.Specular != nil {
			_, err = buff.WriteString(fmt.Sprintf("Ks %g %g %g\n", k.Specular[0], k.Specular[1], k.Specular[2]))
			if err != nil {
				return err
			}
		}
		if k.Emissive != nil {
			_, err = buff.WriteString(fmt.Sprintf("Ke %g %g %g\n", k.Emissive[0], k.Emissive[1], k.Emissive[2]))
			if err != nil {
				return err
			}
		}
		if k.TransmissionFilter != nil {
			_, err = buff.WriteString(fmt.Sprintf("Tf %g %g %g\n", k.TransmissionFilter[0], k.TransmissionFilter[1], k.TransmissionFilter[2]))
			if err != nil {
				return err
			}
		}
		if !math.IsNaN(k.Shininess) {
			_, err = buff.WriteString(fmt.Sprintf("Ns %g\n", k.Shininess))
			if err != nil {
				return err
			}
		}
		if !math.IsNaN(k.Opacity) {
			_, err = buff.WriteString(fmt.Sprintf("d %g\n", k.Opacity))
			if err != nil {
				return err
			}
		}
		if k.AmbientTexture != "" {
			_, err = buff.WriteString(fmt.Sprintf("map_Ka %s\n", k.AmbientTexture))
			if err != nil {
				return err
			}
		}
		if k.DiffuseTexture != "" {
			_, err = buff.WriteString(fmt.Sprintf("map_Kd %s\n", k.DiffuseTexture))
			if err != nil {
				return err
			}
		}
		if k.SpecularTexture != "" {
			_, err = buff.WriteString(fmt.Sprintf("map_Ks %s\n", k.SpecularTexture))
			if err != nil {
				return err
			}
		}
		if k.EmissiveTexture != "" {
			_, err = buff.WriteString(fmt.Sprintf("map_Ke %s\n", k.EmissiveTexture))
			if err != nil {
				return err
			}
		}
		if k.AlphaTexture != "" {
			_, err = buff.WriteString(fmt.Sprintf("map_d %s\n", k.AlphaTexture))
			if err != nil {
				return err
			}
		}
		if k.BumpTexture != "" {
			_, err = buff.WriteString(fmt.Sprintf("map_bump %s\n", k.BumpTexture))
			if err != nil {
				return err
			}
		}
		if k.Illumination != 0 {
			_, err = buff.WriteString(fmt.Sprintf("illum %d\n", k.Illumination))
			if err != nil {
				return err
			}
		}
		if k.Roughness != 0 {
			_, err = buff.WriteString(fmt.Sprintf("Pr %g\n", k.Roughness))
			if err != nil {
				return err
			}
		}
		if k.Metallic != 0 {
			_, err = buff.WriteString(fmt.Sprintf("Pm %g\n", k.Metallic))
			if err != nil {
				return err
			}
		}
		if k.Sheen != 0 {
			_, err = buff.WriteString(fmt.Sprintf("Ps %g\n", k.Sheen))
			if err != nil {
				return err
			}
		}
		if k.ClearcoatThickness != 0 {
			_, err = buff.WriteString(fmt.Sprintf("Pc %g\n", k.ClearcoatThickness))
			if err != nil {
				return err
			}
		}
		if k.ClearcoatRoughness != 0 {
			_, err = buff.WriteString(fmt.Sprintf("Pcr %g\n", k.ClearcoatRoughness))
			if err != nil {
				return err
			}
		}
		if k.Anisotropy != 0 {
			_, err = buff.WriteString(fmt.Sprintf("aniso %g\n", k.Anisotropy))
			if err != nil {
				return err
			}
		}
		if k.AnisotropyRotation != 0 {
			_, err = buff.WriteString(fmt.Sprintf("anisor %g\n", k.AnisotropyRotation))
			if err != nil {
				return err
			}
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buff.Bytes())
	if err != nil {
		return err
	}
	return nil
}

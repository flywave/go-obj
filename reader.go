package obj

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
)

var faceVertexOnlyRegex *regexp.Regexp
var faceVertexAndTexcoordRegex *regexp.Regexp
var faceVertexAndNormalTexcoordRegex *regexp.Regexp
var faceVertexAndNormalRegex *regexp.Regexp
var groupRegex *regexp.Regexp
var usemtlRegex *regexp.Regexp
var mtllibRegex *regexp.Regexp

func init() {
	faceVertexOnlyRegex = regexp.MustCompile(`^(\d+)$`)
	faceVertexAndTexcoordRegex = regexp.MustCompile(`^(\d+)\/(\d+)$`)
	faceVertexAndNormalTexcoordRegex = regexp.MustCompile(`^(\d+)\/(\d+)\/(\d+)$`)
	faceVertexAndNormalRegex = regexp.MustCompile(`^(\d+)\/\/(\d+)$`)
	groupRegex = regexp.MustCompile(`^g\s*(.*)$`)
	usemtlRegex = regexp.MustCompile(`^usemtl\s+(.*)$`)
	mtllibRegex = regexp.MustCompile(`^mtllib\s+(.*)$`)
}

func FirstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

type ObjReader struct {
	ObjBuffer

	options ReadOptions
}

func (l *ObjReader) SetOptions(options ReadOptions) {
	l.options = options
}

func (l *ObjReader) Read(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	i := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		i++
		if hashPos := strings.IndexRune(line, '#'); hashPos != -1 {
			line = line[0:hashPos]
		}
		if len(line) == 0 {
			continue
		}

		var err error
		fields := strings.Fields(line)
		switch strings.ToLower(fields[0]) {
		case "vt":
			err = l.processVertexTexCoord(fields[1:])
		case "v":
			err = l.processVertex(fields[1:])
		case "vn":
			err = l.processVertexNormal(fields[1:])
		case "f":
			err = l.processFace(fields[1:])
		case "g":
			err = l.processGroup(line)
		case "mtllib":
			err = l.processMaterialLibrary(line)
		case "usemtl":
			err = l.processUseMaterial(line)
		case "o":
		case "s":
		case "vp":
			break

		default:
			err = fmt.Errorf("Unknown keyword '%s'", fields[0])
		}

		if err != nil {
			return lineError{i, line, err}
		}
	}
	l.endGroup()
	return scanner.Err()
}

func (l *ObjReader) processVertex(fields []string) error {
	if len(fields) != 3 && len(fields) != 4 {
		return fmt.Errorf("Expected 3 or 4 fields, but got %d", len(fields))
	}
	x, errX := strconv.ParseFloat(fields[0], 32)
	y, errY := strconv.ParseFloat(fields[1], 32)
	z, errZ := strconv.ParseFloat(fields[2], 32)
	if err := FirstError(errX, errY, errZ); err != nil {
		return err
	}
	l.V = append(l.V, vec3.T{float32(x), float32(y), float32(z)})
	return nil
}

func (l *ObjReader) processVertexTexCoord(fields []string) error {
	if len(fields) != 2 {
		return fmt.Errorf("Expected 3 fields, but got %d", len(fields))
	}
	s, errS := strconv.ParseFloat(fields[0], 32)
	t, errT := strconv.ParseFloat(fields[1], 32)
	if err := FirstError(errS, errT); err != nil {
		return err
	}
	l.VT = append(l.VT, vec2.T{float32(s), float32(t)})
	return nil
}

func (l *ObjReader) processVertexNormal(fields []string) error {
	if len(fields) != 3 {
		return fmt.Errorf("Expected 3 fields, but got %d", len(fields))
	}
	x, errX := strconv.ParseFloat(fields[0], 32)
	y, errY := strconv.ParseFloat(fields[1], 32)
	z, errZ := strconv.ParseFloat(fields[2], 32)
	if err := FirstError(errX, errY, errZ); err != nil {
		return err
	}
	l.VN = append(l.VN, vec3.T{float32(x), float32(y), float32(z)})
	return nil
}

func parseFaceField(field string) (faceCorner, error) {
	if match := faceVertexOnlyRegex.FindStringSubmatch(field); match != nil {
		v, err := strconv.Atoi(match[1])
		return faceCorner{v - 1, -1, -1}, err
	} else if match := faceVertexAndTexcoordRegex.FindStringSubmatch(field); match != nil {
		v, errV := strconv.Atoi(match[1])
		t, errN := strconv.Atoi(match[2])
		return faceCorner{v - 1, -1, t - 1}, FirstError(errV, errN)
	} else if match := faceVertexAndNormalTexcoordRegex.FindStringSubmatch(field); match != nil {
		v, errV := strconv.Atoi(match[1])
		t, errN := strconv.Atoi(match[2])
		n, errT := strconv.Atoi(match[3])
		return faceCorner{v - 1, n - 1, t - 1}, FirstError(errV, errN, errT)
	} else if match := faceVertexAndNormalRegex.FindStringSubmatch(field); match != nil {
		v, errV := strconv.Atoi(match[1])
		n, errT := strconv.Atoi(match[2])
		return faceCorner{v - 1, n - 1, -1}, FirstError(errV, errT)
	} else {
		return faceCorner{-1, -1, -1}, fmt.Errorf("Face field '%s' is not on a supported format", field)
	}
}

func (l *ObjReader) isFaceAccepted(f *face) bool {
	if l.options.DiscardDegeneratedFaces {
		occurences := make(map[int]bool, len(f.Corners))
		for _, c := range f.Corners {
			vIdx := c.VertexIndex
			if _, ok := occurences[vIdx]; ok {
				return false
			}
			occurences[vIdx] = true
		}
	}
	return true
}

func (l *ObjReader) processFace(fields []string) error {
	if len(fields) < 3 {
		return fmt.Errorf("Expected %d fields, but got %d", 3, len(fields))
	}

	f := face{make([]faceCorner, len(fields)), l.activeMaterial}
	for i, field := range fields {
		corner, err := parseFaceField(field)
		if err != nil {
			return err
		}
		f.Corners[i] = corner
	}
	if l.isFaceAccepted(&f) {
		l.F = append(l.F, f)
	}
	return nil
}

func (l *ObjReader) processGroup(line string) error {
	if match := groupRegex.FindStringSubmatch(line); match != nil {
		l.endGroup()
		l.startGroup(match[1])
		return nil
	}
	return fmt.Errorf("Could not parse group")
}

func (l *ObjReader) processMaterialLibrary(line string) error {
	if l.MTL != "" {
		return fmt.Errorf("Material library already set")
	}
	if match := mtllibRegex.FindStringSubmatch(line); match != nil {
		l.MTL = match[1]
		return nil
	}
	return fmt.Errorf("Could not parse 'mtllib'-line")
}

func (l *ObjReader) processUseMaterial(line string) error {
	if match := usemtlRegex.FindStringSubmatch(line); match != nil {
		l.activeMaterial = match[1]
		return nil
	}
	return fmt.Errorf("Could not parse 'usemtl'-line")
}

func (l *ObjReader) startGroup(name string) {
	g := group{
		Name:           name,
		FirstFaceIndex: len(l.F),
		FaceCount:      -1,
	}
	l.G = append(l.G, g)
}

func (l *ObjReader) isGroupAccepted(f *face) bool {
	if l.options.DiscardDegeneratedFaces {
		occurences := make(map[int]bool, len(f.Corners))
		for _, c := range f.Corners {
			vIdx := c.VertexIndex
			if _, ok := occurences[vIdx]; ok {
				return false
			}
			occurences[vIdx] = true
		}
	}
	return true
}

func (l *ObjReader) endGroup() {
	if len(l.G) > 0 {
		idx := len(l.G) - 1
		count := len(l.F) - l.G[idx].FirstFaceIndex
		if count > 0 {
			l.G[idx].FaceCount = count
		} else {
			if len(l.G) > 0 {
				l.G = l.G[:len(l.G)-1]
			} else {
				l.G = nil
			}
		}
	} else {
		g := group{
			Name:           "default group",
			FirstFaceIndex: 0,
			FaceCount:      len(l.F),
		}
		l.G = append(l.G, g)
	}
}

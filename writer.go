package obj

import (
	"fmt"
	"io"

	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
)

func (b *ObjBuffer) Write(w io.Writer) error {
	var err error
	_, err = io.WriteString(w,
		fmt.Sprintf("# Exported using RenderDB\n"+
			"# %d vertices, %d normals, %d faces\n",
			len(b.V), len(b.VN), len(b.F)))
	if err != nil {
		return err
	}
	if b.MTL != "" {
		_, err = io.WriteString(w, fmt.Sprintf("mtllib %s\n", b.MTL))
		if err != nil {
			return err
		}
	}
	if err = b.writeVertices(w); err != nil {
		return err
	}
	if err = b.writeNormals(w); err != nil {
		return err
	}
	if err = b.writeTexcoords(w); err != nil {
		return err
	}
	for _, g := range b.G {
		if err = b.writeGroup(w, g); err != nil {
			return err
		}
	}

	return nil
}

func (b *ObjBuffer) writeVertices(w io.Writer) error {
	return writeVectors(w, "v %g %g %g\n", b.V)
}

func (b *ObjBuffer) writeNormals(w io.Writer) error {
	return writeVectors(w, "vn %g %g %g\n", b.VN)
}

func (b *ObjBuffer) writeTexcoords(w io.Writer) error {
	return writeVectors2(w, "vt %g %g\n", b.VT)
}

func writeFace(w io.Writer, f face) error {
	var err error

	_, err = io.WriteString(w, "f")
	if err != nil {
		return err
	}

	for _, c := range f.Corners {
		if c.NormalIndex != -1 {
			if c.TexcoordIndex != -1 {
				_, err = io.WriteString(w,
					fmt.Sprintf(" %d/%d/%d", c.VertexIndex+1, c.TexcoordIndex+1, c.NormalIndex+1))
			} else {
				_, err = io.WriteString(w,
					fmt.Sprintf(" %d//%d", c.VertexIndex+1, c.NormalIndex+1))
			}
		} else if c.TexcoordIndex != -1 {
			_, err = io.WriteString(w,
				fmt.Sprintf(" %d/%d", c.VertexIndex+1, c.TexcoordIndex+1))
		} else {
			_, err = io.WriteString(w, fmt.Sprintf(" %d", c.VertexIndex+1))
		}
		if err != nil {
			return err
		}
	}
	_, err = io.WriteString(w, "\n")
	return err
}

func writeVectors(w io.Writer, format string, vectors []vec3.T) error {
	for _, v := range vectors {
		_, err := io.WriteString(w, fmt.Sprintf(format, v[0], v[1], v[2]))
		if err != nil {
			return err
		}
	}
	return nil
}

func writeVectors2(w io.Writer, format string, vectors []vec2.T) error {
	for _, v := range vectors {
		_, err := io.WriteString(w, fmt.Sprintf(format, v[0], v[1]))
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *ObjBuffer) writeGroup(w io.Writer, g group) error {
	var err error
	_, err = io.WriteString(w, fmt.Sprintf("g %s\n", g.Name))
	if err != nil {
		return err
	}
	for i := g.FirstFaceIndex; i < g.FirstFaceIndex+g.FaceCount; i++ {
		if err = writeFace(w, b.F[i]); err != nil {
			return err
		}
	}

	return nil
}

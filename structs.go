package obj

import (
	"fmt"
	"math"

	"github.com/flywave/go3d/vec2"
	"github.com/flywave/go3d/vec3"
)

type lineError struct {
	lineNumber int
	line       string
	err        error
}

func (e lineError) Error() string {
	return fmt.Sprintf("Line #%d: %v ('%s')", e.lineNumber, e.line, e.err)
}

type faceCorner struct {
	VertexIndex   int
	NormalIndex   int
	TexCoordIndex int
}

type line struct {
	Corners  []int
	Material string
}

type face struct {
	Corners  []faceCorner
	Material string
}

func pnpoly(nvert int, vertx, verty []float32, testx, testy float32) bool {
	i, j := 0, 0
	j = nvert - 1
	c := false
	for i = 0; i < nvert; i++ {
		if ((verty[i] > testy) != (verty[j] > testy)) &&
			(testx <
				(vertx[j]-vertx[i])*(testy-verty[i])/(verty[j]-verty[i])+
					vertx[i]) {
			c = !c
		}
		j = i
	}
	return c
}

func (f *face) Triangulate(V []vec3.T) [][]faceCorner {
	npolys := len(f.Corners)
	if npolys == 3 {
		return [][]faceCorner{f.Corners}
	}

	axes := [2]int{1, 2}
	faces := f.Corners

	var ret [][]faceCorner
	var i1 faceCorner
	i0, i2 := faces[0], faces[1]

	for k := 0; k < npolys; k++ {
		i0 = faces[(k+0)%npolys]
		i1 = faces[(k+1)%npolys]
		i2 = faces[(k+2)%npolys]

		vi0 := i0.VertexIndex
		vi1 := i1.VertexIndex
		vi2 := i2.VertexIndex

		if vi0 >= len(V) || vi1 >= len(V) ||
			vi2 >= len(V) {
			continue
		}
		v0x := V[vi0][0]
		v0y := V[vi0][1]
		v0z := V[vi0][2]
		v1x := V[vi1][0]
		v1y := V[vi1][1]
		v1z := V[vi1][2]
		v2x := V[vi2][0]
		v2y := V[vi2][1]
		v2z := V[vi2][2]
		e0x := v1x - v0x
		e0y := v1y - v0y
		e0z := v1z - v0z
		e1x := v2x - v1x
		e1y := v2y - v1y
		e1z := v2z - v1z
		cx := float32(math.Abs(float64(e0y*e1z - e0z*e1y)))
		cy := float32(math.Abs(float64(e0z*e1x - e0x*e1z)))
		cz := float32(math.Abs(float64(e0x*e1y - e0y*e1x)))
		epsilon := float32(1.19209290e-07)
		if cx > epsilon || cy > epsilon || cz > epsilon {
			if cx > cy && cx > cz {
			} else {
				axes[0] = 0
				if cz > cx && cz > cy {
					axes[1] = 1
				}
			}
			break
		}
	}

	area := float32(0)
	for k := 0; k < npolys; k++ {
		ii0 := faces[(k+0)%npolys]
		ii1 := faces[(k+1)%npolys]
		vi0 := ii0.VertexIndex
		vi1 := ii1.VertexIndex
		if vi0+axes[0] >= len(V) || vi0+axes[1] >= len(V) || vi1+axes[0] >= len(V) || vi1+axes[1] >= len(V) {
			continue
		}
		v0x := V[vi0][axes[0]]
		v0y := V[vi0][axes[1]]
		v1x := V[vi1][axes[0]]
		v1y := V[vi1][axes[1]]
		area += float32(v0x*v1y-v0y*v1x) * 0.5
	}

	maxRounds := 10

	remainingFace := faces
	guessVert := 0
	var ind [3]faceCorner
	var vx [3]float32
	var vy [3]float32

	for len(remainingFace) > 3 && maxRounds > 0 {
		npolys = len(remainingFace)
		if guessVert >= npolys {
			maxRounds--
			guessVert -= npolys
		}
		for k := 0; k < 3; k++ {
			ind[k] = remainingFace[(guessVert+k)%npolys]
			vi := ind[k].VertexIndex
			if vi+axes[0] >= len(V) || vi+axes[1] >= len(V) {
				vx[k] = 0.0
				vy[k] = 0.0
			} else {
				vx[k] = V[vi][axes[0]]
				vy[k] = V[vi][axes[1]]
			}
		}
		e0x := vx[1] - vx[0]
		e0y := vy[1] - vy[0]
		e1x := vx[2] - vx[1]
		e1y := vy[2] - vy[1]
		cross := e0x*e1y - e0y*e1x

		if cross*area < 0.0 {
			guessVert++
			continue
		}

		overlap := false
		for otherVert := 3; otherVert < npolys; otherVert++ {
			idx := (guessVert + otherVert) % npolys

			if idx >= len(remainingFace) {
				continue
			}

			ovi := remainingFace[idx].VertexIndex

			if ((ovi + axes[0]) >= len(V)) ||
				((ovi + axes[1]) >= len(V)) {
				continue
			}
			tx := V[ovi][axes[0]]
			ty := V[ovi][axes[1]]
			if pnpoly(3, vx[:], vy[:], tx, ty) {
				overlap = true
				break
			}
		}

		if overlap {
			guessVert++
			continue
		}

		var idx0, idx1, idx2 faceCorner
		idx0.VertexIndex = ind[0].VertexIndex
		idx0.NormalIndex = ind[0].NormalIndex
		idx0.TexCoordIndex = ind[0].TexCoordIndex
		idx1.VertexIndex = ind[1].VertexIndex
		idx1.NormalIndex = ind[1].NormalIndex
		idx1.TexCoordIndex = ind[1].TexCoordIndex
		idx2.VertexIndex = ind[2].VertexIndex
		idx2.NormalIndex = ind[2].NormalIndex
		idx2.TexCoordIndex = ind[2].TexCoordIndex

		ret = append(ret, []faceCorner{idx0, idx1, idx2})

		removedVertIndex := (guessVert + 1) % npolys
		for removedVertIndex+1 < npolys {
			remainingFace[removedVertIndex] =
				remainingFace[removedVertIndex+1]
			removedVertIndex++
		}
		remainingFace = remainingFace[:len(remainingFace)-1]
	}

	if len(remainingFace) == 3 {
		i0 = remainingFace[0]
		i1 = remainingFace[1]
		i2 = remainingFace[2]

		var idx0, idx1, idx2 faceCorner
		idx0.VertexIndex = i0.VertexIndex
		idx0.NormalIndex = i0.NormalIndex
		idx0.TexCoordIndex = i0.TexCoordIndex
		idx1.VertexIndex = i1.VertexIndex
		idx1.NormalIndex = i1.NormalIndex
		idx1.TexCoordIndex = i1.TexCoordIndex
		idx2.VertexIndex = i2.VertexIndex
		idx2.NormalIndex = i2.NormalIndex
		idx2.TexCoordIndex = i2.TexCoordIndex

		ret = append(ret, []faceCorner{idx0, idx1, idx2})
	}
	return ret
}

type ObjBuffer struct {
	activeMaterial string

	MTL       string
	V         []vec3.T
	VN        []vec3.T
	VT        []vec2.T
	F         []face
	L         []line
	G         []group
	FaceGroup []*faceGroup
}

func (b *ObjBuffer) BoundingBox() vec3.Box {
	box := vec3.Box{Min: vec3.MaxVal, Max: vec3.MinVal}
	for _, v := range b.V {
		box.Join(&vec3.Box{Min: v, Max: v})
	}
	return box
}

type ReadOptions struct {
	DiscardDegeneratedFaces bool
}

package obj

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/flywave/go3d/vec3"
	"github.com/stretchr/testify/assert"
)

func TestObjReader_ProcessMaterialLibrary_InvalidLine_ReturnsError(t *testing.T) {
	loader := ObjReader{}
	assert.Error(t, loader.processMaterialLibrary("invalid mtllib line"))
}

func TestObjReader_ProcessMaterialLibrary_ValidLine_SetsLibrary(t *testing.T) {
	loader := ObjReader{}
	err := loader.processMaterialLibrary("mtllib      materials.mtl")
	assert.NoError(t, err)
	assert.Equal(t, "materials.mtl", loader.MTL)
}

func TestObjReader_ProcessMaterialLibrary_AlreadySet_ReturnsError(t *testing.T) {
	loader := ObjReader{}
	loader.MTL = "somefile.mtl"
	assert.Error(t, loader.processMaterialLibrary("mtllib materials.mtl"))
}

func TestObjReader_ProcessGroup_ValidLine_EndsAndStartsGroup(t *testing.T) {
	// Arrange
	loader := ObjReader{}
	loader.F = []face{face{}}
	loader.G = append(loader.G, group{FirstFaceIndex: 0, FaceCount: -1})

	// Act
	err := loader.processGroup("g   group")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, loader.G[0].FaceCount)
	assert.Equal(t, 2, len(loader.G))
	assert.Equal(t, "group", loader.G[1].Name)
}

func TestObjReader_ProcessGroup_InvalidLine_ReturnsError(t *testing.T) {
	loader := ObjReader{}
	err := loader.processUseMaterial("not a g line")
	assert.Error(t, err)
}

func TestObjReader_ProcessUseMaterial_ValidLine_SetsActiveMaterial(t *testing.T) {
	// Arrange
	loader := ObjReader{}
	loader.F = []face{face{}}

	// Act
	err := loader.processUseMaterial("usemtl       material_name")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "material_name", loader.activeMaterial)
}

func TestObjReader_ProcessFace_InvalidFields_ReturnsError(t *testing.T) {
	loader := ObjReader{}
	assert.Error(t, loader.processFace([]string{}))
	assert.Error(t, loader.processFace([]string{"a", "b", "c"}))
	assert.Error(t, loader.processFace([]string{"1/", "2/", "3/"}))
	assert.Error(t, loader.processFace([]string{"1/1", "2/2", "3/2"})) // Valid but not supported
	assert.Error(t, loader.processFace([]string{"1", "2"}))            // Too few coordinates
}

func TestObjReader_ProcessFace_VertexOnlyFormat_AddsFace(t *testing.T) {
	// Arrange
	loader := ObjReader{}

	// Act
	err := loader.processFace([]string{"1", "2", "3"})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loader.F))
	assert.Equal(t, 3, len(loader.F[0].Corners))
	// Zero-based indices
	assert.Equal(t, 0, loader.F[0].Corners[0].VertexIndex)
	assert.Equal(t, 1, loader.F[0].Corners[1].VertexIndex)
	assert.Equal(t, 2, loader.F[0].Corners[2].VertexIndex)
	assert.Equal(t, -1, loader.F[0].Corners[0].NormalIndex)
	assert.Equal(t, -1, loader.F[0].Corners[1].NormalIndex)
	assert.Equal(t, -1, loader.F[0].Corners[2].NormalIndex)
}

func TestObjReader_ProcessVertex_XYZ_AddsVertex(t *testing.T) {
	// Arrange
	loader := ObjReader{}

	// Act
	err := loader.processVertex([]string{"1.1", "2.0", "3"})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loader.V))
	assert.Equal(t, vec3.T{1.1, 2, 3}, loader.V[0])
}

func TestObjReader_ProcessVertex_XYZW_IgnoresW(t *testing.T) {
	// Arrange
	loader := ObjReader{}

	// Act
	err := loader.processVertex([]string{"1", "2", "3", "999"})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loader.V))
	assert.Equal(t, vec3.T{1, 2, 3}, loader.V[0])
}

func TestObjReader_ProcessVertex_InvalidFields_ReturnsError(t *testing.T) {
	loader := ObjReader{}
	assert.Error(t, loader.processVertex([]string{"0", "0"}))                // XY only
	assert.Error(t, loader.processVertex([]string{"0", "0", "A"}))           // Non-number
	assert.Error(t, loader.processVertex([]string{"0", "0", "0", "1", "2"})) // More than 4 coordinates
}

func TestObjReader_ProcessVertexNormal_XYZ_AddsNormal(t *testing.T) {
	// Arrange
	loader := ObjReader{}

	// Act
	err := loader.processVertexNormal([]string{"1.1", "2.0", "3"})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loader.VN))
	assert.Equal(t, vec3.T{1.1, 2, 3}, loader.VN[0])
}

func TestObjReader_ProcessVertexNormal_InvalidFields_ReturnsError(t *testing.T) {
	loader := ObjReader{}
	assert.Error(t, loader.processVertexNormal([]string{"0", "0"}))           // XY only
	assert.Error(t, loader.processVertexNormal([]string{"0", "0", "A"}))      // Non-number
	assert.Error(t, loader.processVertexNormal([]string{"0", "0", "0", "1"})) // More than 3 coordinates
}

func TestObjReader_StartGroup_StartsNewGroup(t *testing.T) {
	// Arrange
	loader := ObjReader{}

	// Act
	loader.startGroup("MyGroup")

	// Assert
	assert.Equal(t, 1, len(loader.G))
	assert.Equal(t, "MyGroup", loader.G[0].Name)
	assert.Equal(t, 0, loader.G[0].FirstFaceIndex)
	assert.Equal(t, -1, loader.G[0].FaceCount)
}

func TestObjReader_EndGroup_NoGroups_DoesNotPanic(t *testing.T) {
	loader := ObjReader{}
	assert.NotPanics(t, func() {
		loader.endGroup()
	})
}

func TestObjReader_EndGroup_GroupStarted_UpdatesFaceCount(t *testing.T) {
	// Arrange
	loader := ObjReader{}
	loader.G = append(loader.G, group{
		Name:           "Test",
		FirstFaceIndex: 0,
		FaceCount:      -1,
	})

	// Act
	loader.F = append(loader.F, createFace("mat", 1, 2, 3))
	loader.endGroup()

	// Assert
	assert.Equal(t, []group{group{"Test", 0, 1}}, loader.G)
}

func TestObjReader_ProcessFace_UsesActiveMaterial(t *testing.T) {
	// Arrange
	loader := ObjReader{}
	loader.activeMaterial = "my-material"

	// Act
	err := loader.processFace([]string{"1", "2", "3"})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(loader.F))
	assert.Equal(t, "my-material", loader.F[0].Material)
}

func TestObjReader_EndGroup_EmptyGroup_DiscardsLast(t *testing.T) {
	// Arrange
	loader := ObjReader{}
	origGroups := []group{group{Name: "first"}}
	loader.G = origGroups

	// Act
	loader.startGroup("last")
	loader.endGroup()

	// Assert
	assert.EqualValues(t, origGroups, loader.G)
}

func TestLoadObj(t *testing.T) {
	loader := ObjReader{}
	file, err := os.Open("../data/test.obj")
	if err != nil {
		t.Error(err)
	}

	err = loader.Read(file)
	if err != nil {
		t.Error(err)
	}
}

func TestLoadLineObj(t *testing.T) {
	loader := ObjReader{}
	file, err := os.Open("./line.obj")
	if err != nil {
		t.Error(err)
	}

	err = loader.Read(file)
	if err != nil {
		t.Error(err)
	}

	wr := [][][3]float64{}
	for i := range loader.L {
		c := loader.L[i].Corners
		p1 := loader.V[c[0]]
		p2 := loader.V[c[1]]

		wr = append(wr, [][3]float64{
			[3]float64{float64(p1[0]), float64(p1[1]), float64(p1[2])},
			[3]float64{float64(p2[0]), float64(p2[1]), float64(p2[2])},
		})
	}

	data, _ := json.Marshal(wr)

	os.WriteFile("./line.json", data, os.ModePerm)

}

func TestBBox(t *testing.T) {
	loader := ObjReader{}
	file, err := os.Open("./1_8.obj")
	if err != nil {
		t.Error(err)
	}

	err = loader.Read(file)
	if err != nil {
		t.Error(err)
	}

	bbox := loader.BoundingBox()
	center := bbox.Center()

	for i, v := range loader.V {
		loader.V[i] = vec3.Sub(&v, &center)
	}

	f, _ := os.Create("./1_8_2.obj")

	loader.Write(f)
	f.Close()
}

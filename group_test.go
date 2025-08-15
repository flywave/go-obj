package obj

import (
	"testing"

	"github.com/flywave/go3d/vec3"
	"github.com/stretchr/testify/assert"
)

func createFace(material string, cornerIdx ...int) Face {
	f := Face{}
	f.Corners = make([]FaceCorner, len(cornerIdx))
	for i := 0; i < len(cornerIdx); i++ {
		f.Corners[i].VertexIndex = cornerIdx[i]
		f.Corners[i].NormalIndex = cornerIdx[i]
	}
	f.Material = material
	return f
}

func TestGroup_BuildFormats_EmptyGroup_ReturnsEmptyBuffer(t *testing.T) {
	// Arrange
	g := Group{}
	origBuffer := ObjBuffer{}
	origBuffer.MTL = "materials.mtl"

	// Act
	buffer := g.buildBuffers(&origBuffer)

	// Assert
	assert.Equal(t, "materials.mtl", buffer.MTL)
	assert.Equal(t, 0, len(buffer.F))
	assert.Equal(t, 0, len(buffer.V))
	assert.Equal(t, 0, len(buffer.VN))
}

func TestGroup_BuildFormats_SingleGroupWithSingleFace_ReturnsCorrect(t *testing.T) {
	// Arrange
	g := Group{}
	g.FirstFaceIndex = 0
	g.FaceCount = 1

	origBuffer := ObjBuffer{}
	origBuffer.G = []Group{g}
	origBuffer.F = []Face{
		createFace("mat", 0, 1, 2),
	}
	origBuffer.V = []vec3.T{
		{0, 0, 0},
		{1, 1, 1},
		{2, 2, 2},
	}
	origBuffer.VN = []vec3.T{
		{0, 0, 0},
		{-1, -1, -1},
		{-2, -2, -2},
	}

	// Act
	buffer := g.buildBuffers(&origBuffer)

	// Assert
	assert.Equal(t, 1, len(buffer.G))
	assert.Equal(t, 1, len(buffer.F))
	assert.Equal(t, 3, len(buffer.V))
	assert.Equal(t, 3, len(buffer.VN))
}

func TestGroup_BuildFormats_TwoGroupsWithTwoFaces_ReturnsCorrectGroups(t *testing.T) {
	// Arrange
	origBuffer := ObjBuffer{}
	origBuffer.F = []Face{
		// Group 1
		createFace("mat1", 0, 2, 4),
		createFace("mat2", 4, 2, 6),
		// Group 2
		createFace("mat1", 1, 3, 5),
		createFace("mat2", 5, 3, 7),
	}
	origBuffer.V = []vec3.T{
		{0, 0, 0},
		{1, 1, 1},
		{2, 2, 2},
		{3, 3, 3},
		{4, 4, 4},
		{5, 5, 5},
		{6, 6, 6},
		{7, 7, 7},
	}
	origBuffer.VN = []vec3.T{
		{0, 0, 0},
		{-1, -1, -1},
		{-2, -2, -2},
		{-3, -3, -3},
		{-4, -4, -4},
		{-5, -5, -5},
		{-6, -6, -6},
		{-7, -7, -7},
	}

	g1 := Group{Name: "Group 1", FirstFaceIndex: 0, FaceCount: 2}
	g2 := Group{Name: "Group 2", FirstFaceIndex: 2, FaceCount: 2}
	origBuffer.G = []Group{g1, g2}

	// Act
	buffer := g1.buildBuffers(&origBuffer)

	// Assert
	assert.EqualValues(t,
		[]vec3.T{
			{0, 0, 0}, {2, 2, 2}, {4, 4, 4}, {6, 6, 6},
		},
		buffer.V)
	assert.EqualValues(t,
		[]vec3.T{
			{0, 0, 0}, {-2, -2, -2}, {-4, -4, -4}, {-6, -6, -6},
		},
		buffer.VN)
	assert.Equal(t, 1, len(buffer.G))
	assert.Equal(t,
		Group{Name: "Group 1", FirstFaceIndex: 0, FaceCount: 2},
		buffer.G[0])
	assert.Equal(t, 2, len(buffer.F))
	assert.Equal(t, "mat1", buffer.F[0].Material)
	assert.Equal(t, "mat2", buffer.F[1].Material)
}

func TestGroup_BuildFormats_GroupWithTwoFacesets_ReturnsCorrectSubset(t *testing.T) {
	// Arrange
	origBuffer := ObjBuffer{}
	origBuffer.F = []Face{
		// Group 1
		createFace("Material 1", 0, 2, 4),
		createFace("Material 1", 4, 2, 6),
		createFace("Material 2", 1, 3, 5),
		createFace("Material 2", 5, 3, 4),
		// Group 2
		createFace("Material 3", 5, 7, 2),
		createFace("Material 3", 7, 5, 4),
	}
	origBuffer.V = []vec3.T{
		{0, 0, 0},
		{1, 1, 1},
		{2, 2, 2},
		{3, 3, 3},
		{4, 4, 4},
		{5, 5, 5},
		{6, 6, 6},
		{7, 7, 7},
	}
	origBuffer.VN = []vec3.T{
		{0, 0, 0},
		{-1, -1, -1},
		{-2, -2, -2},
		{-3, -3, -3},
		{-4, -4, -4},
		{-5, -5, -5},
		{-6, -6, -6},
		{-7, -7, -7},
	}

	g1 := Group{Name: "Group 1", FirstFaceIndex: 0, FaceCount: 4}
	g2 := Group{Name: "Group 2", FirstFaceIndex: 4, FaceCount: 2}
	origBuffer.G = []Group{g1, g2}

	// Act
	buffer := g2.buildBuffers(&origBuffer)

	// Assert
	assert.EqualValues(t,
		[]vec3.T{
			{5, 5, 5}, {7, 7, 7}, {2, 2, 2}, {4, 4, 4},
		},
		buffer.V)
	assert.EqualValues(t,
		[]vec3.T{
			{-5, -5, -5}, {-7, -7, -7}, {-2, -2, -2}, {-4, -4, -4},
		},
		buffer.VN)
	assert.EqualValues(t, []Face{
		createFace("Material 3", 0, 1, 2), // Remapped indices
		createFace("Material 3", 1, 0, 3), // Remapped indices
	}, buffer.F)
	assert.EqualValues(t, []Group{{"Group 2", 0, 2}}, buffer.G)
}

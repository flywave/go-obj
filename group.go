package obj

func FillIntSlice(slice []int, val int) {
	for i := 0; i < len(slice); i++ {
		slice[i] = val
	}
}

type group struct {
	Name           string
	FirstFaceIndex int
	FaceCount      int
}

func (g *group) buildBuffers(parentBuffer *ObjBuffer) *ObjBuffer {
	buffer := new(ObjBuffer)
	buffer.MTL = parentBuffer.MTL
	buffer.G = []group{
		group{
			Name:      g.Name,
			FaceCount: g.FaceCount,
		},
	}

	vertexMapping := make([]int, len(parentBuffer.V))
	FillIntSlice(vertexMapping, -1)
	normalMapping := make([]int, len(parentBuffer.VN))
	FillIntSlice(normalMapping, -1)

	for i := g.FirstFaceIndex; i < g.FirstFaceIndex+g.FaceCount; i++ {

		originalFace := parentBuffer.F[i]

		f := face{Material: originalFace.Material}
		f.Corners = make([]faceCorner, len(originalFace.Corners))

		for j, origCorner := range originalFace.Corners {
			origVertIdx := origCorner.VertexIndex
			origNormIdx := origCorner.NormalIndex

			var newVertIdx int
			if newVertIdx = vertexMapping[origVertIdx]; newVertIdx == -1 {
				newVertIdx = len(buffer.V)
				buffer.V = append(buffer.V, parentBuffer.V[origVertIdx])
				vertexMapping[origVertIdx] = newVertIdx
			}

			var newNormIdx int
			if newNormIdx = normalMapping[origNormIdx]; newNormIdx == -1 {
				newNormIdx = len(buffer.VN)
				buffer.VN = append(buffer.VN, parentBuffer.VN[origNormIdx])
				normalMapping[origNormIdx] = newNormIdx
			}

			f.Corners[j].VertexIndex, f.Corners[j].NormalIndex = newVertIdx, newNormIdx
		}

		buffer.F = append(buffer.F, f)
	}
	return buffer
}

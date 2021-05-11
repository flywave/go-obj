package obj

import "testing"

func TestMaterial(t *testing.T) {
	mtls, err := ReadMaterials("../data/test.mtl")

	if err != nil {
		t.Error(err)
	}

	if len(mtls) == 0 {
		t.Error("error")
	}
}

package object

import (
	"testing"
)

func TestStringHashKey(t *testing.T) {
	hash1 := &String{Value: "Hello World"}
	hash2 := &String{Value: "Hello World"}
	hash3 := &String{Value: "My name is johnny"}
	hash4 := &String{Value: "My name is johnny"}

	if hash1.HashKey() != hash2.HashKey() {
		t.Errorf("strings with same content but have different hash keys")
	}

	if hash3.HashKey() != hash4.HashKey() {
		t.Errorf("strings with same content but have different hash keys")
	}

	if hash1.HashKey() == hash3.HashKey() {
		t.Errorf("strings with different content but have same hash keys")
	}
}

func TestIntegerHashKey(t *testing.T) {
	hash1 := &Integer{Value: 1}
	hash2 := &Integer{Value: 1}
	hash3 := &Integer{Value: 2}
	hash4 := &Integer{Value: 2}

	if hash1.HashKey() != hash2.HashKey() {
		t.Errorf("integers with same content but have different hash keys")
	}

	if hash3.HashKey() != hash4.HashKey() {
		t.Errorf("integers with same content but have different hash keys")
	}

	if hash1.HashKey() == hash3.HashKey() {
		t.Errorf("integers with different content but have same hash keys")
	}
}

func TestBooleanHashKey(t *testing.T) {
	hash1 := &Boolean{Value: true}
	hash2 := &Boolean{Value: true}
	hash3 := &Boolean{Value: false}
	hash4 := &Boolean{Value: false}

	if hash1.HashKey() != hash2.HashKey() {
		t.Errorf("booleans with same content but have different hash keys")
	}

	if hash3.HashKey() != hash4.HashKey() {
		t.Errorf("booleans with same content but have different hash keys")
	}

	if hash1.HashKey() == hash3.HashKey() {
		t.Errorf("boolean with different content but have same hash keys")
	}
}

package blowfish

// #cgo pkg-config: libcrypto
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <openssl/blowfish.h>
// #include <stdlib.h>
import "C"
import "unsafe"

type BFKey struct {
	key *C.BF_KEY
}

func NewBFKey(data []byte) *BFKey {
	k := &BFKey{key: (*C.BF_KEY)(C.malloc(C.size_t(unsafe.Sizeof(C.BF_KEY{}))))}
	C.BF_set_key(k.key, C.int(len(data)), (*C.uchar)(unsafe.Pointer(&data[0])))
	return k
}

func (k *BFKey) Free() {
	C.free(unsafe.Pointer(k.key))
}

type Blowfish struct {
	key *BFKey
}

func New(key []byte) *Blowfish {
	k := NewBFKey(key)
	return &Blowfish{key: k}
}

func (b *Blowfish) BlockSize() int {
	return 8
}

func (b *Blowfish) Encrypt(dst, src []byte) {
	if len(src) < 8 || len(dst) < 8 {
		panic("BF_ecb_encrypt requires 8-byte blocks")
	}
	C.BF_ecb_encrypt(
		(*C.uchar)(unsafe.Pointer(&src[0])),
		(*C.uchar)(unsafe.Pointer(&dst[0])),
		b.key.key,
		C.int(1),
	)
}

func (b *Blowfish) Decrypt(dst, src []byte) {
	if len(src) < 8 || len(dst) < 8 {
		panic("BF_ecb_encrypt requires 8-byte blocks")
	}
	C.BF_ecb_encrypt(
		(*C.uchar)(unsafe.Pointer(&src[0])),
		(*C.uchar)(unsafe.Pointer(&dst[0])),
		b.key.key,
		C.int(0),
	)
}

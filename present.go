// Package present is an implementation of the PRESENT lightweight block cipher
package present

/*
   Mechanical translation of https://github.com/michaelkitson/Present-8bit
*/

const (
	ROUNDS                    = 32
	ROUND_KEY_SIZE_BYTES      = 8
	PRESENT_BLOCK_SIZE_BYTES  = 8
	PRESENT_80_KEY_SIZE_BYTES = 10
)

var sBox = [16]byte{
	0xC, 0x5, 0x6, 0xB, 0x9, 0x0, 0xA, 0xD, 0x3, 0xE, 0xF, 0x8, 0x4, 0x7, 0x1, 0x2}

var sBoxInverse = [16]byte{
	0x5, 0xE, 0xF, 0x8, 0xC, 0x1, 0x2, 0xD, 0xB, 0x4, 0x6, 0x3, 0x0, 0x7, 0x9, 0xA}

func copyKey(from, to []byte) {
	copy(to, from)
}

func generateRoundKeys80(suppliedKey []byte, keys *[ROUNDS][ROUND_KEY_SIZE_BYTES]byte) {
	// trashable key copies
	var key [PRESENT_80_KEY_SIZE_BYTES]byte
	var newKey [PRESENT_80_KEY_SIZE_BYTES]byte
	var i, j byte
	copyKey(suppliedKey[:], key[:])
	copyKey(key[:], keys[0][:])
	for i = 1; i < ROUNDS; i++ {
		// rotate left 61 bits
		for j = 0; j < PRESENT_80_KEY_SIZE_BYTES; j++ {
			newKey[j] = (key[(j+7)%PRESENT_80_KEY_SIZE_BYTES] << 5) |
				(key[(j+8)%PRESENT_80_KEY_SIZE_BYTES] >> 3)
		}
		copyKey(newKey[:], key[:])

		// pass leftmost 4-bits through sBox
		key[0] = (sBox[key[0]>>4] << 4) | (key[0] & 0xF)

		// xor roundCounter into bits 15 through 19
		key[8] ^= i << 7 // bit 15
		key[7] ^= i >> 1 // bits 19-16

		copyKey(key[:], keys[i][:])
	}
}

func addRoundKey(block []byte, roundKey *[8]byte) {
	var i byte
	for i = 0; i < PRESENT_BLOCK_SIZE_BYTES; i++ {
		block[i] ^= roundKey[i]
	}
}

func pLayer(block []byte) {
	var i, j, indexVal, andVal byte
	var initial [PRESENT_BLOCK_SIZE_BYTES]byte
	copyKey(block[:], initial[:])
	for i = 0; i < PRESENT_BLOCK_SIZE_BYTES; i++ {
		block[i] = 0
		for j = 0; j < 8; j++ {
			indexVal = 4*(i%2) + (3 - (j >> 1))
			andVal = (8 >> (i >> 1)) << ((j % 2) << 2)
			block[i] |= bool2byte((initial[indexVal]&andVal) != 0) << j
		}
	}
}

func pLayerInverse(block []byte) {
	var i, j, indexVal, andVal byte
	var initial [PRESENT_BLOCK_SIZE_BYTES]byte
	copyKey(block[:], initial[:])
	for i = 0; i < PRESENT_BLOCK_SIZE_BYTES; i++ {
		block[i] = 0
		for j = 0; j < 8; j++ {
			indexVal = (7 - ((2 * j) % 8)) - bool2byte(i < 4)
			andVal = (7 - ((2 * i) % 8)) - bool2byte(j < 4)
			block[i] |= bool2byte((initial[indexVal]&(1<<andVal)) != 0) << j
		}
	}
}

func Present80_encryptBlock(block []byte, key []byte) {
	var roundKeys [ROUNDS][ROUND_KEY_SIZE_BYTES]byte
	var i, j byte
	generateRoundKeys80(key, &roundKeys)
	for i = 0; i < ROUNDS-1; i++ {
		addRoundKey(block, &roundKeys[i])
		for j = 0; j < PRESENT_BLOCK_SIZE_BYTES; j++ {
			block[j] = (sBox[block[j]>>4] << 4) | sBox[block[j]&0xF]
		}
		pLayer(block)
	}
	addRoundKey(block, &roundKeys[ROUNDS-1])
}

func Present80_decryptBlock(block []byte, key []byte) {
	var roundKeys [ROUNDS][ROUND_KEY_SIZE_BYTES]byte
	var i, j byte
	generateRoundKeys80(key, &roundKeys)
	for i = ROUNDS - 1; i > 0; i-- {
		addRoundKey(block, &roundKeys[i])
		pLayerInverse(block)
		for j = 0; j < PRESENT_BLOCK_SIZE_BYTES; j++ {
			block[j] = (sBoxInverse[block[j]>>4] << 4) | sBoxInverse[block[j]&0xF]
		}
	}
	addRoundKey(block, &roundKeys[0])
}

func bool2byte(b bool) byte {
	if b {
		return 1
	}

	return 0
}

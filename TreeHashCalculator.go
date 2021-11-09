package main

import (
  "crypto/sha256"
)

// Splits data into {chunkSize} chunks, and calculate the final hash by
// combining the hash of each chunk
func ComputeSHA256TreeHash(data []byte, chunkSize int) []byte {
  hashes := GetHashesChunks(data, chunkSize)
  return ComputeCombineHashChunks(hashes)
}

func GetHashesChunks(data []byte, chunkSize int) [][]byte {
  dataLen := len(data)
  size := CeilQuotient(dataLen, chunkSize)
  res := make([][]byte, size)

  i, off := 0, 0
  for off < dataLen {
    lo := off
    hi := Min(lo + chunkSize, dataLen)

    copy := data[lo:hi]
    hash := sha256.Sum256(copy)
    res[i] = hash[:]

    off = hi
    i = i + 1
  }

  return res
}

func ComputeCombineHashChunks(hashes [][]byte) []byte {
  for len(hashes) > 1 {
    tmp := make([][]byte, CeilQuotient(len(hashes), 2))

    for i, j := 0, 0; i < len(hashes); i, j = i+2, j+1 {
      a := hashes[i]
      b := []byte{}
      if i+1 < len(hashes) {
        b = hashes[i+1]
      }

      combined := append(a, b...)
      hash := sha256.Sum256(combined)
      tmp[j] = hash[:]
    }

    hashes = tmp
  }

  return hashes[0]
}
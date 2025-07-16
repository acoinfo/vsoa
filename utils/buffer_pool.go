// Copyright (c) 2023 ACOAUTO Team.
// All rights reserved.
//
// Detailed license information can be found in the LICENSE file.
//
// File: buffer_pool.go Vehicle SOA utils package.
//
// Author: Wang.yifan <wangyifan@acoinfo.com>
// Contributor: Cheng.siyuan <chengsiyuan@acoinfo.com>

package utils

import (
	"sync"
)

// Constants for findPool operation types
const (
	POOL_GET = iota
	POOL_PUT
)

// LimitedPool is a memory pool for managing byte slices of different sizes.
type LimitedPool struct {
	minSize int
	maxSize int
	pools   []*levelPool
	// Precomputed index mapping for quick lookup
	indexMap []int
}

// levelPool is an internal structure of LimitedPool, used to manage byte slices of a specific size.
type levelPool struct {
	size int
	pool sync.Pool
}

// NewLimitedPool creates a new LimitedPool instance.
// minSize is the minimum size of the byte slices, and maxSize is the maximum size.
func NewLimitedPool(minSize, maxSize int) *LimitedPool {
	if maxSize < minSize {
		panic("maxSize can't be less than minSize")
	}
	const multiplier = 2
	var pools []*levelPool
	curSize := minSize
	for curSize <= maxSize {
		pools = append(pools, newLevelPool(curSize))
		curSize *= multiplier
	}

	// Precompute index mapping for quick lookup
	indexMap := make([]int, maxSize+1)
	for i, pool := range pools {
		for size := pool.size; size <= maxSize; size++ {
			if indexMap[size] == 0 {
				indexMap[size] = i
			}
		}
	}

	return &LimitedPool{
		minSize:  minSize,
		maxSize:  maxSize,
		pools:    pools,
		indexMap: indexMap,
	}
}

// newLevelPool creates a new levelPool instance.
func newLevelPool(size int) *levelPool {
	return &levelPool{
		size: size,
		pool: sync.Pool{
			New: func() any {
				data := make([]byte, size)
				return &data
			},
		},
	}
}

// findPool finds the appropriate levelPool based on the requested size and operation type.
// opType can be GET or PUT.
func (p *LimitedPool) findPool(size int, opType int) *levelPool {
	if size < p.minSize || size > p.maxSize {
		return nil
	}
	idx := p.indexMap[size]
	if idx == 0 && size != p.minSize {
		if opType == POOL_GET {
			// For Get, find the next available pool
			for i := 1; i < len(p.indexMap); i++ {
				if p.indexMap[i] != 0 {
					idx = p.indexMap[i]
					break
				}
			}
		} else {
			return nil
		}
	}
	return p.pools[idx]
}

// Get retrieves a byte slice of the specified size from the memory pool.
// If no suitable pool is found, a new slice is allocated directly.
func (p *LimitedPool) Get(size int) *[]byte {
	sp := p.findPool(size, POOL_GET)
	if sp == nil {
		data := make([]byte, size)
		return &data
	}
	buf := sp.pool.Get().(*[]byte)
	*buf = (*buf)[:size]
	return buf
}

// Put recycles a byte slice back into the memory pool.
// If no suitable pool is found, the slice is discarded.
func (p *LimitedPool) Put(b *[]byte) {
	sp := p.findPool(cap(*b), POOL_PUT)
	if sp == nil {
		return
	}
	*b = (*b)[:cap(*b)]
	// Clear the slice content to avoid memory fragmentation
	for i := range *b {
		(*b)[i] = 0
	}
	sp.pool.Put(b)
}

func ResizeSliceSize[T ~[]byte](b T, size int) T {
	if cap(b) < size {
		return make(T, size)
	}
	return b[:size]
}

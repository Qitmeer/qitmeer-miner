package kernel

var CuckarooKernelNew = `
// Cuckaroo Cycle, a memory-hard proof-of-work by James Qitmeer
// Copyright (c) 2019 
// edgebits {{edge_bits}}

#pragma OPENCL EXTENSION cl_khr_int64_base_atomics : enable
#pragma OPENCL EXTENSION cl_khr_int64_extended_atomics : enable

typedef uint8 u8;
typedef uint16 u16;
typedef uint u32;
typedef ulong u64;

typedef u32 node_t;
typedef u64 nonce_t;

#define EDGEBITS {{edge_bits}}
#define STEP {{step}}
// number of edges
#define NEDGES ((node_t)1 << EDGEBITS)
// used to mask siphash output
#define EDGEMASK (NEDGES - 1)

#define SIPROUND \
  do { \
    v0 += v1; v2 += v3; v1 = rotate(v1,(ulong)13); \
    v3 = rotate(v3,(ulong)16); v1 ^= v0; v3 ^= v2; \
    v0 = rotate(v0,(ulong)32); v2 += v1; v0 += v3; \
    v1 = rotate(v1,(ulong)17);   v3 = rotate(v3,(ulong)21); \
    v1 ^= v2; v3 ^= v0; v2 = rotate(v2,(ulong)32); \
  } while(0)

constant u8 zero = 0;

static u32 dipnode(ulong v0, ulong v1, ulong v2, ulong v3, u64 nce, bool uorv) {
		v3 ^= nce;
		for (int r = 0; r < 2; r++)
			SIPROUND;
		v0 ^= nce ;
		v2 ^= 0xff;
		for (int r = 0; r < 4; r++)
			SIPROUND;

		u64 v = (v0 ^ v1) ^ (v2  ^ v3);	
		return uorv ? ((( ( v ) & EDGEMASK)<<1) | 1) : (( v & EDGEMASK)<<1);
}


__attribute__((reqd_work_group_size({{group}}, 1, 1)))
__kernel  void CreateEdges(const u64 v0i, const u64 v1i, const u64 v2i, const u64 v3i, __global uchar * edges,__global u32 * indexes0,__global u32 * indexes1)
{
	const int gid = get_global_id(0);
	__global u32 * indexes;
	u32 index;
	for (int i = 0; i < STEP; i += 1)
	{
		u64 blockNonce = gid * STEP + i;
		u32 u = dipnode(v0i,v1i,v2i,v3i,(blockNonce << 1),false);
		u32 v = dipnode(v0i,v1i,v2i,v3i,(blockNonce << 1 | 1),true);
		//u64 index = u+V;
			
			edges[blockNonce] = 1;
			indexes = u < NEDGES ? indexes0 : indexes1;
			index = u < NEDGES ? u : u - NEDGES;
			atomic_inc(&indexes[index]);
			indexes = v < NEDGES ? indexes0 : indexes1;
			index = v < NEDGES ? v : v - NEDGES;
			atomic_inc(&indexes[index]);
	}

}

__attribute__((reqd_work_group_size({{group}}, 1, 1)))
__kernel  void Trimmer01(const u64 v0i, const u64 v1i, const u64 v2i, const u64 v3i,__global uchar * edges,__global u32 *indexes0,__global u32 *indexes1)
{
	const int gid = get_global_id(0);
	__global u32 * indexesU;
	__global u32 * indexesV;
	u32 indexU;
	u32 indexV;
	for (int i = 0; i < STEP; i++)
	{
		u32 blockNonce = gid * STEP + i;
		u32 u = dipnode(v0i,v1i,v2i,v3i,(blockNonce << 1),false);
		u32 v = dipnode(v0i,v1i,v2i,v3i,(blockNonce << 1 | 1),true);
		if(edges[blockNonce]==0){
			continue;
		}
		indexesU = u < NEDGES ? indexes0 : indexes1;
		indexU = u < NEDGES ? u : u - NEDGES;
		indexesV = v < NEDGES ? indexes0 : indexes1;
		indexV = v < NEDGES ? v : v - NEDGES;
		if(indexesU[indexU]==1 && indexesV[indexV]>1){
			atomic_dec(&indexesV[indexV]);
			edges[blockNonce] = 0;
		}
		if(indexesV[indexV]==1 && indexesU[indexU]>1){	
			atomic_dec(&indexesU[indexU]);
			edges[blockNonce] = 0;
		}
	}
	
}


__attribute__((reqd_work_group_size({{group}}, 1, 1)))
__kernel  void Trimmer02(const u64 v0i, const u64 v1i, const u64 v2i, const u64 v3i,__global uchar * edges,__global u32 *indexes0,__global u32 *indexes1,__global u32 * destination,__global u32 *count)
{
	const int gid = get_global_id(0);
	__global u32 * indexesU;
	__global u32 * indexesV;
	u32 indexU;
	u32 indexV;
	for (int i = 0; i < STEP; i++)
	{
		u64 blockNonce = gid * STEP + i;
		u32 u = dipnode(v0i,v1i,v2i,v3i,(blockNonce << 1),false);
		u32 v = dipnode(v0i,v1i,v2i,v3i,(blockNonce << 1 | 1),true);
		
		if(edges[blockNonce]==0){
			continue;
		}
		indexesU = u < NEDGES ? indexes0 : indexes1;
		indexU = u < NEDGES ? u : u - NEDGES;
		indexesV = v < NEDGES ? indexes0 : indexes1;
		indexV = v < NEDGES ? v : v - NEDGES;
		if(indexesU[indexU]>1 && indexesV[indexV]>1){
			int idx = atomic_add(&count[0],1);
			atomic_add(&count[1],1);
			destination[idx] = blockNonce;
		}
	}
	
}

`
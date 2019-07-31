package kernel

var CuckarooKernel = `
// Cuckaroo Cycle, a memory-hard proof-of-work by James Qitmeer
// Copyright (c) 2019 

#pragma OPENCL EXTENSION cl_khr_int64_base_atomics : enable
#pragma OPENCL EXTENSION cl_khr_int64_extended_atomics : enable

typedef uint8 u8;
typedef uint16 u16;
typedef uint u32;
typedef ulong u64;

typedef u32 node_t;
typedef u64 nonce_t;

#define EDGEBITS 24
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

__attribute__((reqd_work_group_size(256, 1, 1)))
__kernel  void CreateEdges(const u64 v0i, const u64 v1i, const u64 v2i, const u64 v3i, __global u32 * edges,__global u32 * indexes)
{
	const int gid = get_global_id(0);

	u64 u00;
	u64 v00;

	u64 v0;
	u64 v1;
	u64 v2;
	u64 v3;

	for (int i = 0; i < 16; i += 1)
	{
		u64 blockNonce = gid * 16 + i;
		u64 nonce1 = (blockNonce << 1);
		u64 nonce2 = (blockNonce << 1 | 1);
		//build u
		v0 = v0i;
		v1 = v1i;
		v2 = v2i;
		v3 = v3i;

		v3 ^= nonce1;
		for (int r = 0; r < 2; r++)
			SIPROUND;
		v0 ^= nonce1 ;
		v2 ^= 0xff;
		for (int r = 0; r < 4; r++)
			SIPROUND;

		u00 = (v0 ^ v1) ^ (v2  ^ v3);	

		//build V
		v0 = v0i;
		v1 = v1i;
		v2 = v2i;
		v3 = v3i;

		v3 ^= nonce2;
		for (int r = 0; r < 2; r++)
			SIPROUND;
		v0 ^= nonce2 ;
		v2 ^= 0xff;
		for (int r = 0; r < 4; r++)
			SIPROUND;

		v00 = (v0 ^ v1) ^ (v2  ^ v3);	
		u32 u = (( u00 & EDGEMASK)<<1);
		//u32 V = ((( ( v00 >> 32 ) & EDGEMASK)<<1) | 1);
		u32 V = ((( ( v00 ) & EDGEMASK)<<1) | 1);
		//u64 index = u+V;
		//int idx = atomic_inc(&existBucket[index]);
			edges[nonce1] = u;
			edges[nonce2] = V;
			atomic_inc(&indexes[u]);
			atomic_inc(&indexes[V]);
		//if(idx==0){
		//	edges[nonce1] = u;
		//	edges[nonce2] = V;
		//	atomic_inc(&indexes[u]);
		//	atomic_inc(&indexes[V]);
		//} else{
		//	edges[nonce1] = 0;
		//	edges[nonce2] = 0;
		//}
	}

}

__attribute__((reqd_work_group_size(256, 1, 1)))
__kernel  void Trimmer01(__global uint2 * edges,__global u32 *indexes)
{
	const int gid = get_global_id(0);
	for (int i = 0; i < 16; i++)
	{
		u32 blockNonce = gid * 16 + i;
		u32 V = edges[blockNonce].x;
		u32 v1 = edges[blockNonce].y;
		if(V==0 && v1==0){
			continue;
		}
		if(indexes[V]==1 && indexes[v1]>1){
			atomic_dec(&indexes[v1]);
			edges[blockNonce] = 0;
		}
		if(indexes[v1]==1 && indexes[V]>1){	
			atomic_dec(&indexes[V]);
			edges[blockNonce] = 0;
		}
	}
	
}

__attribute__((reqd_work_group_size(256, 1, 1)))
__kernel  void Trimmer02(__global uint2 * edges,__global u32 *indexes,__global uint2 * destination,__global u32 *count)
{
	const int gid = get_global_id(0);
	barrier(CLK_LOCAL_MEM_FENCE);
	for (int i = 0; i < 16; i++)
	{
		u64 blockNonce = gid * 16 + i;
		u32 V = edges[blockNonce].x;
		u32 v1 = edges[blockNonce].y;
		
		if(V==0 && v1==0){
			continue;
		}

		if(indexes[V]>1 && indexes[v1]>1){

			int idx = atomic_add(&count[0],1);
			atomic_add(&count[1],1);

			destination[idx] = edges[blockNonce];
		}
	}
	
}

__attribute__((reqd_work_group_size(256, 1, 1)))
__kernel  void RecoveryNonce(__global uint2 * edges,__global uint2 *nodes,__global u32 *nonces)
{
	const int gid = get_global_id(0);
	barrier(CLK_LOCAL_MEM_FENCE);
	for (int i = 0; i < 16; i++)
	{
		u64 blockNonce = gid * 16 + i;
		u32 V = edges[blockNonce].x;
		u32 v1 = edges[blockNonce].y;
		
		if(V==0 && v1==0){
			continue;
		}

		for(int k=0;k<42;k++){
			u32 x = nodes[k].x;
			u32 y = nodes[k].y;
			if((V==x &&v1==y) || (V==y &&v1==x) ){
				nonces[k] = blockNonce;
			}
		}
	}
	
}


`

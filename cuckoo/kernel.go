package cuckoo

var CuckarooKernel = `
#pragma OPENCL EXTENSION cl_khr_int64_base_atomics : enable
#pragma OPENCL EXTENSION cl_khr_int64_extended_atomics : enable
typedef uint8 u8;
typedef uint16 u16;
typedef uint u32;
typedef ulong u64;
typedef u32 node_t;
typedef u64 nonce_t;
#define DUCK_SIZE_A 129L
#define DUCK_SIZE_B 83L
#define DUCK_A_EDGES (DUCK_SIZE_A * 1024L)
#define DUCK_A_EDGES_64 (DUCK_A_EDGES * 64L)
#define DUCK_B_EDGES (DUCK_SIZE_B * 1024L)
#define DUCK_B_EDGES_64 (DUCK_B_EDGES * 64L)
#define EDGE_BLOCK_SIZE (64)
#define EDGE_BLOCK_MASK (EDGE_BLOCK_SIZE - 1)
#define EDGEBITS 29
// number of edges
#define NEDGES ((node_t)1 << EDGEBITS)
// used to mask siphash output
#define EDGEMASK (NEDGES - 1)
#define CTHREADS 1024
#define BKTMASK4K (4096-1)
#define BKTGRAN 32
#define SIPROUND \
  do { \
    v0 += v1; v2 += v3; v1 = rotate(v1,(ulong)13); \
    v3 = rotate(v3,(ulong)16); v1 ^= v0; v3 ^= v2; \
    v0 = rotate(v0,(ulong)32); v2 += v1; v0 += v3; \
    v1 = rotate(v1,(ulong)17);   v3 = rotate(v3,(ulong)21); \
    v1 ^= v2; v3 ^= v0; v2 = rotate(v2,(ulong)32); \
  } while(0)
void Increase2bCounter(__local u32 * ecounters, const int bucket)
{
	int word = bucket >> 5;
	unsigned char bit = bucket & 0x1F;
	u32 mask = 1 << bit;
	u32 old = atomic_or(ecounters + word, mask) & mask;
	if (old > 0)
		atomic_or(ecounters + word + 4096, mask);
}
bool Read2bCounter(__local u32 * ecounters, const int bucket)
{
	int word = bucket >> 5;
	unsigned char bit = bucket & 0x1F;
	u32 mask = 1 << bit;
	return (ecounters[word + 4096] & mask) > 0;
}
__attribute__((reqd_work_group_size(128, 1, 1)))
__kernel  void FluffySeed2A(const u64 v0i, const u64 v1i, const u64 v2i, const u64 v3i, __global ulong4 * bufferA, __global ulong4 * bufferB, __global u32 * indexes)
{
	const int gid = get_global_id(0);
	const short lid = get_local_id(0);
	__global ulong4 * buffer;
	__local u64 tmp[64][16];
	__local u32 counters[64];
	u64 sipblock[64];
	u64 v0;
	u64 v1;
	u64 v2;
	u64 v3;
	if (lid < 64)
		counters[lid] = 0;

	barrier(CLK_LOCAL_MEM_FENCE);
	for (int i = 0; i < 1024 * 2; i += EDGE_BLOCK_SIZE)
	{
		u64 blockNonce = gid * (1024 * 2) + i;
		v0 = v0i;
		v1 = v1i;
		v2 = v2i;
		v3 = v3i;
		for (u32 b = 0; b < EDGE_BLOCK_SIZE; b++)
		{
			v3 ^= blockNonce + b;
			for (int r = 0; r < 2; r++)
				SIPROUND;
			v0 ^= blockNonce + b;
			v2 ^= 0xff;
			for (int r = 0; r < 4; r++)
				SIPROUND;
			sipblock[b] = (v0 ^ v1) ^ (v2  ^ v3);
		}
		u64 last = sipblock[EDGE_BLOCK_MASK];
		for (short s = 0; s < EDGE_BLOCK_SIZE; s++)
		{
			u64 lookup = s == EDGE_BLOCK_MASK ? last : sipblock[s] ^ last;
			uint2 hash = (uint2)(lookup & EDGEMASK, (lookup >> 32) & EDGEMASK);
			int bucket = hash.x & 63;
			barrier(CLK_LOCAL_MEM_FENCE);
			int counter = atomic_add(counters + bucket, (u32)1);
			int counterLocal = counter % 16;
			tmp[bucket][counterLocal] = hash.x | ((u64)hash.y << 32);
			barrier(CLK_LOCAL_MEM_FENCE);
			
			if ((counter > 0) && (counterLocal == 0 || counterLocal == 8))
			{
				int cnt = min((int)atomic_add(indexes + bucket, 8), (int)(DUCK_A_EDGES_64 - 8));
				int idx = ((bucket < 32 ? bucket : bucket - 32) * DUCK_A_EDGES_64 + cnt) / 4;
				buffer = bucket < 32 ? bufferA : bufferB;
				buffer[idx] = 1232142535;
				return;
				//buffer[idx] = (ulong4)(
				//	atom_xchg(&tmp[bucket][8 - counterLocal], (u64)0),
				//	atom_xchg(&tmp[bucket][9 - counterLocal], (u64)0),
				//	atom_xchg(&tmp[bucket][10 - counterLocal], (u64)0),
				//	atom_xchg(&tmp[bucket][11 - counterLocal], (u64)0)
				//);
				//buffer[idx + 1] = (ulong4)(
				//	atom_xchg(&tmp[bucket][12 - counterLocal], (u64)0),
				//	atom_xchg(&tmp[bucket][13 - counterLocal], (u64)0),
				//	atom_xchg(&tmp[bucket][14 - counterLocal], (u64)0),
				//	atom_xchg(&tmp[bucket][15 - counterLocal], (u64)0)
				//);
			}
		}
	}
	
}

`

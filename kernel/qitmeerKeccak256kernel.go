package kernel

const QitmeerKeccak256kernelSource = `
/*
 * Scrypt-jane public domain, OpenCL implementation of scrypt(keccak,chacha,SCRYPTN,1,1) 2013 mtrlt
 * Modify By Qitmeer Team 2020-09-07
 * Mining Qitmeer - 0.9.2
 */

// kernel-interface: fullheader Keccak

uint2 ROTL64_1(const uint2 x, const uint y);
uint2 ROTL64_2(const uint2 x, const uint y);
#define ARGS_25(x) x ## 0, x ## 1, x ## 2, x ## 3, x ## 4, x ## 5, x ## 6, x ## 7, x ## 8, x ## 9, x ## 10, x ## 11, x ## 12, x ## 13, x ## 14, x ## 15, x ## 16, x ## 17, x ## 18, x ## 19, x ## 20, x ## 21, x ## 22, x ## 23, x ## 24

void keccak_block_noabsorb(ARGS_25(uint2* s));
__constant uint2 keccak_round_constants[24] = {
	(uint2)(0x00000001,0x00000000), (uint2)(0x00008082,0x00000000),
	(uint2)(0x0000808a,0x80000000), (uint2)(0x80008000,0x80000000),
	(uint2)(0x0000808b,0x00000000), (uint2)(0x80000001,0x00000000),
	(uint2)(0x80008081,0x80000000), (uint2)(0x00008009,0x80000000),
	(uint2)(0x0000008a,0x00000000), (uint2)(0x00000088,0x00000000),
	(uint2)(0x80008009,0x00000000), (uint2)(0x8000000a,0x00000000),
	(uint2)(0x8000808b,0x00000000), (uint2)(0x0000008b,0x80000000),
	(uint2)(0x00008089,0x80000000), (uint2)(0x00008003,0x80000000),
	(uint2)(0x00008002,0x80000000), (uint2)(0x00000080,0x80000000),
	(uint2)(0x0000800a,0x00000000), (uint2)(0x8000000a,0x80000000),
	(uint2)(0x80008081,0x80000000), (uint2)(0x00008080,0x80000000),
	(uint2)(0x80000001,0x00000000), (uint2)(0x80008008,0x80000000)
};

uint2 ROTL64_1(const uint2 x, const uint y)
{
	return (uint2)((x.x<<y)^(x.y>>(32-y)),(x.y<<y)^(x.x>>(32-y)));
}
uint2 ROTL64_2(const uint2 x, const uint y)
{
	return (uint2)((x.y<<y)^(x.x>>(32-y)),(x.x<<y)^(x.y>>(32-y)));
}

#define RND(i) \
do{  \
		m0 = *s0 ^ *s5 ^ *s10 ^ *s15 ^ *s20 ^ ROTL64_1(*s2 ^ *s7 ^ *s12 ^ *s17 ^ *s22, 1);\
		m1 = *s1 ^ *s6 ^ *s11 ^ *s16 ^ *s21 ^ ROTL64_1(*s3 ^ *s8 ^ *s13 ^ *s18 ^ *s23, 1);\
		m2 = *s2 ^ *s7 ^ *s12 ^ *s17 ^ *s22 ^ ROTL64_1(*s4 ^ *s9 ^ *s14 ^ *s19 ^ *s24, 1);\
		m3 = *s3 ^ *s8 ^ *s13 ^ *s18 ^ *s23 ^ ROTL64_1(*s0 ^ *s5 ^ *s10 ^ *s15 ^ *s20, 1);\
		m4 = *s4 ^ *s9 ^ *s14 ^ *s19 ^ *s24 ^ ROTL64_1(*s1 ^ *s6 ^ *s11 ^ *s16 ^ *s21, 1);\
\
		m5 = *s1^m0;\
\
		*s0 ^= m4;\
		*s1 = ROTL64_2(*s6^m0, 12);\
		*s6 = ROTL64_1(*s9^m3, 20);\
		*s9 = ROTL64_2(*s22^m1, 29);\
		*s22 = ROTL64_2(*s14^m3, 7);\
		*s14 = ROTL64_1(*s20^m4, 18);\
		*s20 = ROTL64_2(*s2^m1, 30);\
		*s2 = ROTL64_2(*s12^m1, 11);\
		*s12 = ROTL64_1(*s13^m2, 25);\
		*s13 = ROTL64_1(*s19^m3,  8);\
		*s19 = ROTL64_2(*s23^m2, 24);\
		*s23 = ROTL64_2(*s15^m4, 9);\
		*s15 = ROTL64_1(*s4^m3, 27);\
		*s4 = ROTL64_1(*s24^m3, 14);\
		*s24 = ROTL64_1(*s21^m0,  2);\
		*s21 = ROTL64_2(*s8^m2, 23);\
		*s8 = ROTL64_2(*s16^m0, 13);\
		*s16 = ROTL64_2(*s5^m4, 4);\
		*s5 = ROTL64_1(*s3^m2, 28);\
		*s3 = ROTL64_1(*s18^m2, 21);\
		*s18 = ROTL64_1(*s17^m1, 15);\
		*s17 = ROTL64_1(*s11^m0, 10);\
		*s11 = ROTL64_1(*s7^m1,  6);\
		*s7 = ROTL64_1(*s10^m4,  3);\
		*s10 = ROTL64_1(      m5,  1);\
		\
		m5 = *s0; m6 = *s1; *s0 = bitselect(*s0^*s2,*s0,*s1); *s1 = bitselect(*s1^*s3,*s1,*s2); *s2 = bitselect(*s2^*s4,*s2,*s3); *s3 = bitselect(*s3^m5,*s3,*s4); *s4 = bitselect(*s4^m6,*s4,m5);\
		m5 = *s5; m6 = *s6; *s5 = bitselect(*s5^*s7,*s5,*s6); *s6 = bitselect(*s6^*s8,*s6,*s7); *s7 = bitselect(*s7^*s9,*s7,*s8); *s8 = bitselect(*s8^m5,*s8,*s9); *s9 = bitselect(*s9^m6,*s9,m5);\
		m5 = *s10; m6 = *s11; *s10 = bitselect(*s10^*s12,*s10,*s11); *s11 = bitselect(*s11^*s13,*s11,*s12); *s12 = bitselect(*s12^*s14,*s12,*s13); *s13 = bitselect(*s13^m5,*s13,*s14); *s14 = bitselect(*s14^m6,*s14,m5);\
		m5 = *s15; m6 = *s16; *s15 = bitselect(*s15^*s17,*s15,*s16); *s16 = bitselect(*s16^*s18,*s16,*s17); *s17 = bitselect(*s17^*s19,*s17,*s18); *s18 = bitselect(*s18^m5,*s18,*s19); *s19 = bitselect(*s19^m6,*s19,m5);\
		m5 = *s20; m6 = *s21; *s20 = bitselect(*s20^*s22,*s20,*s21); *s21 = bitselect(*s21^*s23,*s21,*s22); *s22 = bitselect(*s22^*s24,*s22,*s23); *s23 = bitselect(*s23^m5,*s23,*s24); *s24 = bitselect(*s24^m6,*s24,m5);\
\
		*s0 ^= keccak_round_constants[i];  \
}while(0)

void keccak_block_noabsorb(ARGS_25(uint2* s))
{
	uint2 m0,m1,m2,m3,m4,m5,m6;
#pragma unroll
	for (uint i = 0; i < 24; ++i)
		RND(i);
}

ulong uint2ToUlong(uint2 x) {
	uchar *a = (uchar *) &x;
	uchar b[8] = {a[7],a[6],a[5],a[4],a[3],a[2],a[1],a[0]};
	ulong *res1 = (ulong *) b;
	ulong hash0 = as_ulong(as_uchar8(*res1).s76543210);
	return hash0;
}

// __attribute__((reqd_work_group_size(256, 1, 1)))
__kernel void search(__global const uint2*restrict in, __global uint*restrict xnonce,__global uint2 *target)
{
	uint2 ARGS_25(state);
	uint nonce = get_global_id(0) + 0xF0000000;
	state0 = in[0];
	state1 = in[1];
	state2 = in[2];
	state3 = in[3];
	state4 = in[4];
	state5 = in[5];
	state6 = in[6];
	state7 = in[7];
	state8 = in[8];
	state9 = in[9];
	state10 = in[10];
	state11 = in[11];
	state12 = in[12];
	// state13 = in[13];
	state13 = (uint2)(in[13].x, nonce);
	state14 = in[14];
	state15 = in[15];
	state16 = (uint2)(0,0xf1000000U);
	state17 = 0;
	state18 = 0;
	state19 = 0;
	state20 = 0;
	state21 = 0;
	state22 = 0;
	state23 = 0;
	state24 = 0;
	
	keccak_block_noabsorb(ARGS_25(&state));
	if(state3.y < target[3].y){
		*xnonce = nonce;
		return;
	}
	if(state3.y == 0 && state3.x < target[3].x){
		*xnonce = nonce;
		return;
	}
	if(state3.y == 0 && state3.x == 0 && state2.y < target[2].y){
		*xnonce = nonce;
		return;
	}
	if(state3.y == 0 && state3.x == 0 && state2.y == 0 && state2.x < target[2].x){
		*xnonce = nonce;
		return;
	}
	if(state3.y == 0 && state3.x == 0 && state2.y == 0 && state2.x == 0 && state1.y < target[1].y){
		*xnonce = nonce;
		return;
	}
	if(state3.y == 0 && state3.x == 0 && state2.y == 0 && state2.x == 0 && state1.y == 0 && state1.x < target[1].x){
		*xnonce = nonce;
		return;
	}
	if(state3.y == 0 && state3.x == 0 && state2.y == 0 && state2.x == 0 && state1.y == 0 && state1.x == 0 && state0.y < target[0].y){
		*xnonce = nonce;
		return;
	}
	if(state3.y == 0 && state3.x == 0 && state2.y == 0 && state2.x == 0 && state1.y == 0 && state1.x == 0 && state0.y == 0 && state0.x < target[0].x){
		*xnonce = nonce;
		return;
	}
}

`

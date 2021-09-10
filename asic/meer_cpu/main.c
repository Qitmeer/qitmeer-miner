#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>


#include "meer.h"

uint8_t* header;


//测试主程序
int meer_cpu_calc(uint8_t* header,uint8_t* target,uint8_t *nonce)
{
    uint32_t accepts = 0;
    uint32_t rejects = 0;
    struct timeval time_prev;
    gettimeofday(&time_prev, NULL);
    uint64_t n = 0;
    while(1) {
        volatile int interval = 0;
        bool matched = false;
        uint8_t hash_out[32]={0};
        n++;
        for(int i=0;i<8;i++) {
            header[109+i] = nonce[i];
            for(int i=0;i<117;i++) {
                printf("%02x", work_temp.header[i]);
            }
            printf("\n");
            meer_hash(hash_out, (uint8_t*)(header));	//计算返回nonce hash值
            printf("target cmp:\n");
            for(int i=0;i<32;i++) {
                 printf("%02x", target[i]);
            }
            printf("\n");
            for(int i=0;i<32;i++) {
                printf("%02x", hash_out[i]);
            }
            printf("\n");
            for(int i=0;i<32;i++) {
                        if(hash_out[31-i] < target[31-i]) {
                            accepts++;
                            matched = true;
                            printf("target matched!");
                            break;
                        }
            }
            if(!matched) {
                rejects++;
            }
            struct timeval time_now;
            gettimeofday(&time_now, NULL);
            int64_t diffone = 0x000000FFFFFFFFFF;
            float diffone_f = (float)diffone;
            int duration = time_now.tv_sec - time_prev.tv_sec;
            if(duration <= 0) {
                 duration = 1;
            }
            printf("Running %d Seconds, accept %d, reject %d, MHS %.2f GH/S, Reject rate %0.2f\n", duration, accepts, rejects, accepts*1.0f*diffone_f/duration/1000000000, ((float)rejects)/((float)(accepts+rejects)));
        }
     }
}

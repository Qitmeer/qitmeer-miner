#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>


#include "uart.h"
#include "meer_drv.h"
#include "meer.h"

#define MEER_DRV_VERSION	"0.1"
#define NUM_OF_CHIPS    1

//辅助函数
static const int hex2bin_tbl[256] = {
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	 0,  1,  2,  3,  4,  5,  6,  7,  8,  9, -1, -1, -1, -1, -1, -1,
	-1, 10, 11, 12, 13, 14, 15, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, 10, 11, 12, 13, 14, 15, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
};

bool hex2bin(unsigned char *p, const char *hexstr, size_t len)
{
	int nibble1, nibble2;
	unsigned char idx;
	bool ret = false;
    
	while (*hexstr && len) {
		if (!hexstr[1]) {
			printf("%s,hex2bin str truncated\n", __func__);			
			return ret;
		}		
        
		idx = *hexstr++;
		nibble1 = hex2bin_tbl[idx];
		
		idx = *hexstr++;
		nibble2 = hex2bin_tbl[idx];

		if ((nibble1 < 0) || (nibble2 < 0)) {
			printf("%s,hex2bin scan failed %d,%d\n", __func__,nibble1, nibble2);			
			return ret;
		}

		*p++ = (((unsigned char)nibble1) << 4) | ((unsigned char)nibble2);
		--len;
	}
	
	if ((len == 0 && *hexstr == 0)) {
		
		ret = true;
	}
	return ret;
}



//测试主程序
int main(int argc, char* argv[])
{
    int fd;
    uint8_t target[32] = {0};
    struct work work_temp;    
    uint8_t header[117]= {0};
	
	printf("meer driver %s\n", MEER_DRV_VERSION);
	
    
	//初始化算力板
    if(meer_drv_init(&fd, NUM_OF_CHIPS)) {
        return -1;
    }

    /*meer_drv_set_freq(fd, 100);
    usleep(500000);
    meer_drv_set_freq(fd, 200);
    usleep(500000);
    meer_drv_set_freq(fd, 300);
    usleep(500000);
    meer_drv_set_freq(fd, 400);
    usleep(500000);
    meer_drv_set_freq(fd, 500);
    usleep(500000);*/

    
    char * ptarget_str = "0000000000000000000000000000000000000000000000000000ffff00000000"; //diff 1
    hex2bin(target, ptarget_str, sizeof(target));
    memcpy(work_temp.target, target, 32); //难度目标配置
    
    char * pheader_str = "1200000003c60b43da920ae08be3dd91e174fc7b5d538ca5601a4ea9fbcfc703447dd4871b7fac4e54a887df6c1801f4ac37883d6808cb93855f1f07aa4c2cfa73eea3b1000000000000000000000000000000000000000000000000000000000000000000f5231c83cf1060080000000000000000";    
    hex2bin(header, pheader_str, sizeof(header));
    memcpy(work_temp.header, header, 117); //meer区块头
    
    meer_drv_set_work(fd, &work_temp, NUM_OF_CHIPS); //对算力板下任务
    struct timeval time_prev;
    gettimeofday(&time_prev, NULL);
    while(1) {
        uint8_t nonce[8];
        uint8_t chip_id;
        uint8_t job_id;
        uint32_t accepts = 0;
        uint32_t rejects = 0;
        bool matched = false;
        if(get_nonce(fd, nonce, &chip_id, &job_id)) {	//读取nonce
            if (1/*(chip_id >= 1) && (chip_id <= NUM_OF_CHIPS)*/) {
                uint8_t hash_out[32]={0};
                for(int i=0;i<8;i++) {
                    work_temp.header[109+i] = nonce[i];
                }
                printf("header in:\n");
                for(int i=0;i<117;i++) {
                    printf("%02x", work_temp.header[i]);
                }
                printf("\n");
                meer_hash(hash_out, (uint8_t*)(work_temp.header));	//计算返回nonce hash值
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
                        break;
                    }
                }
                if(!matched) {
                    rejects++;
                }
                struct timeval time_now;
                gettimeofday(&time_now, NULL);
                int64_t diffone = 0x00000000FFFFFFFF;
                float diffone_f = (float)diffone;
                int duration = time_now.tv_sec - time_prev.tv_sec;
                if(duration <= 0) {
                    duration = 1;
                }
                printf("Running %d Seconds, MHS %.2f GH/S, Reject rate %0.2f\n", duration, accepts*1.0f*diffone_f/duration/1000000000, ((float)rejects)/((float)(accepts+rejects)));
            }
        }
        usleep(10000);
    }

    meer_drv_deinit(fd);
	return 0;
}

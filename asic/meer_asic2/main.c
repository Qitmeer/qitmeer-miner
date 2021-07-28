#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>
#include <signal.h> // signal functions

#include "uart.h"
#include "meer_drv.h"
#include "meer.h"

#define MEER_DRV_VERSION	"0.2asic"
#define NUM_OF_CHIPS    14
#define DEF_WORK_INTERVAL   60000 //ms

int testwork(int fd,int sec);
void getX(int fd);

int pinlv = 0;

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
int clearwork(int fd);
int flag = 0;
static void my_handler(int sig){ // can be called asynchronously
  flag = 1; // set flag
}

//测试主程序
int main(int argc, char* argv[])
{
    if( argc != 3 )
    {
      printf("ERROR Params\n");
      return -1;
    }
    printf("\nUART PATH: %s | GPIO : %s \n",argv[1],argv[2]);
    int fd;
    uint8_t target[32] = {0};
    struct work work_temp;    
    uint8_t header[117]= {0};
    signal(SIGINT, my_handler);
	printf("========meer driver %s\n", MEER_DRV_VERSION);

	//初始化算力板
    if(meer_drv_init(&fd, NUM_OF_CHIPS,argv[1],argv[2])) {
        return -1;
    }
    meer_drv_set_freq(fd, 100);
    usleep(500000);
    uart_write_register(fd,0x90,0x00,0x00,0xff,0x00);   //门控
    usleep(100000);
    uart_write_register(fd,0x90,0x00,00,0x57,0x01);   //group 1
    usleep(100000);
    uart_write_register(fd,0x90,0x00,00,0x58,0x01);   //group 2
    usleep(100000);
    uart_write_register(fd,0x90,0x00,00,0x59,0x01);   //group 3
    usleep(100000);
    uart_write_register(fd,0x90,0x00,0x00,0xff,0x01);
    usleep(100000);
    getX(fd);
    uart_read_register(fd, 0x01, 0x57);
    uart_read_register(fd, 0x01, 0x58);
    uart_read_register(fd, 0x01, 0x59);
    printf("\n =======READ CHIP========\n");
    pinlv = 100;
    meer_drv_set_freq(fd, 125);
    usleep(1000000);
    meer_drv_set_freq(fd, 150);
    usleep(1000000);
    meer_drv_set_freq(fd, 175);
    usleep(1000000);
    meer_drv_set_freq(fd, 200);
    usleep(1000000);
    pinlv = 200;
    meer_drv_set_freq(fd, 225);
    usleep(1000000);
    meer_drv_set_freq(fd, 250);
    usleep(1000000);
    pinlv = 250;
    meer_drv_set_freq(fd, 275);
    usleep(1000000);
    getX(fd);
    usleep(1000000);
    pinlv = 275;
    meer_drv_set_freq(fd, 300);
    getX(fd);
    usleep(1000000);
    pinlv = 300;
    meer_drv_set_freq(fd, 320);
    getX(fd);
    usleep(1000000);
    pinlv = 320;
    meer_drv_set_freq(fd, 325);
    usleep(1000000);
    pinlv = 325;
//    testwork(fd,60);
    meer_drv_set_freq(fd, 331);
    usleep(1000000);
    pinlv = 331;
//    testwork(fd,120);
    meer_drv_set_freq(fd, 337);
    usleep(1000000);
    pinlv = 337;
//    testwork(fd,120);
    meer_drv_set_freq(fd, 343);
    usleep(1000000);
    pinlv = 343;
//    testwork(fd,120);
    meer_drv_set_freq(fd, 350);
    getX(fd);
    usleep(1000000);
    pinlv = 350;
    testwork(fd,600);
    printf("\n **************************exit mining meer_drv_deinit************************** \n");
    meer_drv_deinit(fd,argv[2]);
	return 0;
}

int testwork(int fd,int sec)
{
    uint8_t target[32] = {0};
    struct work work_temp;
    uint8_t header[117]= {0};

    char * ptarget_str = "0000000000000000000000000000000000000000000000000000ff0000000000"; //diff 1
    hex2bin(target, ptarget_str, sizeof(target));
    memcpy(work_temp.target, target, 32); //难度目标配置

    char * pheader_str = "1200000003c60b43da920ae08be3dd91e174fc7b5d538ca5601a4ea9fbcfc703447dd4871b7fac4e54a887df6c1801f4ac37883d6808cb93855f1f07aa4c2cfa73eea3b1000000000000000000000000000000000000000000000000000000000000000000f5231c83cf1060080000000000000000";
    hex2bin(header, pheader_str, sizeof(header));
    memcpy(work_temp.header, header, 117); //meer区块头

    uint32_t accepts = 0;
    uint32_t rejects = 0;
    struct timeval time_prev;
    gettimeofday(&time_prev, NULL);
    while(1) {
        uint8_t nonce[8];
        uint8_t chip_id;
        uint8_t job_id;
        struct timeval time1;
        gettimeofday(&time1, NULL);
        if(time1.tv_sec-time_prev.tv_sec>=sec){
            break;
        }
        if (flag==1){
            break;
        }
        meer_drv_set_work(fd, &work_temp, NUM_OF_CHIPS); //对算力板下任务
        volatile int interval = 0;
        printf("\n ============current pinlv:%d Mhz\n",pinlv);
        while(interval < DEF_WORK_INTERVAL/10) {
            bool matched = false;
             if (flag==1){
                        break;
                    }
            if(get_nonce(fd, nonce, &chip_id, &job_id)) {	//读取nonce
if ((chip_id >= 1) && (chip_id <= NUM_OF_CHIPS)) {
    uint8_t hash_out[32]={0};
    for(int i=0;i<8;i++) {
        work_temp.header[109+i] = nonce[i];
    }
    meer_hash(hash_out, (uint8_t*)(work_temp.header));	//计算返回nonce hash值
    for(int i=0;i<32;i++) {
        if(hash_out[31-i] < target[31-i]) {
            accepts++;
            matched = true;
            printf("\n*******************************target matched!\n");
            printf("\n Hash:: ");
            for(int i=0;i<32;i++) {
printf("%02x", hash_out[i]);
            }
            printf("\n");
            break;
        }
        if(hash_out[31-i] >target[31-i]){
             printf("\n Not Match Hash:: ");
             for(int i=0;i<32;i++) {
 printf("%02x", hash_out[i]);
             }
             printf("\n");
            break;
        }
    }
    if(!matched) {
        rejects++;
        printf("\n pinlv:%d Mhz CHIPID: %d *********************** ERROR CHECK *********************** \n",pinlv,chip_id);
//        clearwork(fd);
//        int oldPinlv = pinlv;
//        pinlv = get_last_freq_reg_data(pinlv);
//        meer_drv_set_freq(fd, pinlv,NUM_OF_CHIPS);
//        usleep(60000000);
//        testwork(fd,120);
//        if(oldPinlv == 100){
//            break;
//        }
//        meer_drv_set_freq(fd, oldPinlv,NUM_OF_CHIPS);
//        usleep(60000000);
//        gettimeofday(&time_prev, NULL);
    }
    struct timeval time_now;
    gettimeofday(&time_now, NULL);
    int64_t diffone = 0x0000000000FFFFFF;
    float diffone_f = (float)diffone;
    int duration = time_now.tv_sec - time_prev.tv_sec;
    if(duration <= 0) {
        duration = 1;
    }
    printf("\npinlv:%d Mhz ***********************Running %d Seconds, accept %d, reject %d, MHS %.2f GH/S, Reject rate %0.2f\n", pinlv,duration, accepts, rejects, accepts*1.0f*diffone_f/duration/1000000000, ((float)rejects)/((float)(accepts+rejects)));
}
            }
            usleep(10000);
            interval++;
        }
    }
//    clearwork(fd);
	return 0;
}

int clearwork(int fd)
{
    printf("\n ********************************** stop stask ********************************** \n");
    uint8_t target[32] = {0};
    struct work work_temp;
    uint8_t header[117]= {0};

    char * ptarget_str = "0000000000000000000000000000000000000000000000000000000000000000"; //diff 1
    hex2bin(target, ptarget_str, sizeof(target));
    memcpy(work_temp.target, target, 32); //难度目标配置

    char * pheader_str = "1200000003c60b43da920ae08be3dd91e174fc7b5d538ca5601a4ea9fbcfc703447dd4871b7fac4e54a887df6c1801f4ac37883d6808cb93855f1f07aa4c2cfa73eea3b1000000000000000000000000000000000000000000000000000000000000000000f5231c83cf1060080000000000000000";
    hex2bin(header, pheader_str, sizeof(header));
    memcpy(work_temp.header, header, 117); //meer区块头
    meer_drv_set_work(fd, &work_temp, NUM_OF_CHIPS); //对算力板下任务
	return 0;
}

void getX(int fd){
    uart_read_register(fd, 0x01, 0x00);
    uart_read_register(fd, 0x02, 0x00);
    uart_read_register(fd, 0x03, 0x00);
    uart_read_register(fd, 0x04, 0x00);
    uart_read_register(fd, 0x05, 0x00);
    uart_read_register(fd, 0x06, 0x00);
    uart_read_register(fd, 0x07, 0x00);
    uart_read_register(fd, 0x08, 0x00);
    uart_read_register(fd, 0x09, 0x00);
    uart_read_register(fd, 0x0a, 0x00);
    uart_read_register(fd, 0x0b, 0x00);
    uart_read_register(fd, 0x0c, 0x00);
    uart_read_register(fd, 0x0d, 0x00);
    uart_read_register(fd, 0x0e, 0x00);
}

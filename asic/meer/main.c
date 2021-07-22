#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>

#include "uart.h"
#include "meer_drv.h"
#include "meer.h"
#include "main.h"
#define MEER_DRV_VERSION	"0.2asic"
#define NUM_OF_CHIPS	1
#define DEF_WORK_INTERVAL   30000 //msx


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
int init_drv(int num_of_chips)
{
	int fd;
	printf("\n****************meer driver %s\n", MEER_DRV_VERSION);
	//初始化算力板
	if(meer_drv_init(&fd, num_of_chips)) {
		return -1;
	}

	 meer_drv_set_freq(fd, 100);
        usleep(500000);
        printf("\n =======READ CHIP========\n");
        uart_read_register(fd, 0x01, 0x00);
        meer_drv_set_freq(fd, 125);
        usleep(500000);
        meer_drv_set_freq(fd, 150);
        usleep(500000);
        meer_drv_set_freq(fd, 175);
        usleep(500000);
        meer_drv_set_freq(fd, 200);
        usleep(500000);
        meer_drv_set_freq(fd, 225);
        usleep(500000);
        meer_drv_set_freq(fd, 250);
        usleep(500000);
        meer_drv_set_freq(fd, 275);
        usleep(500000);
        meer_drv_set_freq(fd, 300);
        usleep(500000);
        meer_drv_set_freq(fd, 320);
        usleep(500000);
        meer_drv_set_freq(fd, 325);
        usleep(500000);
        meer_drv_set_freq(fd, 331);
        usleep(500000);
        meer_drv_set_freq(fd, 337);
        usleep(500000);
        meer_drv_set_freq(fd, 337);
        usleep(500000);
        meer_drv_set_freq(fd, 343);
        usleep(500000);
        meer_drv_set_freq(fd, 350);
        usleep(500000);
        uart_read_register(fd, 0x01, 0x00);
        meer_drv_set_freq(fd, 356);
        usleep(500000);
        meer_drv_set_freq(fd, 362);
        usleep(500000);
        meer_drv_set_freq(fd, 368);
        usleep(500000);
        meer_drv_set_freq(fd, 375);
        usleep(500000);
        meer_drv_set_freq(fd, 381);
        usleep(500000);
        meer_drv_set_freq(fd, 387);
        usleep(500000);
        meer_drv_set_freq(fd, 393);
        usleep(500000);
        meer_drv_set_freq(fd, 400);
        usleep(500000);
        meer_drv_set_freq(fd, 425);
        usleep(500000);
        meer_drv_set_freq(fd, 445);
        usleep(500000);
        meer_drv_set_freq(fd, 447);
        usleep(500000);
        meer_drv_set_freq(fd, 450);
        usleep(500000);
        meer_drv_set_freq(fd, 452);
        usleep(500000);
        meer_drv_set_freq(fd, 455);
        usleep(500000);
        meer_drv_set_freq(fd, 457);
        usleep(500000);
        meer_drv_set_freq(fd, 460);
        usleep(500000);
        meer_drv_set_freq(fd, 462);
        usleep(500000);
        meer_drv_set_freq(fd, 465);
        usleep(500000);
        meer_drv_set_freq(fd, 467);
        usleep(500000);
        meer_drv_set_freq(fd, 470);
        usleep(500000);
        meer_drv_set_freq(fd, 472);
        usleep(500000);
        meer_drv_set_freq(fd, 475);
        usleep(500000);
        meer_drv_set_freq(fd, 477);
        usleep(500000);
        meer_drv_set_freq(fd, 480);
        usleep(500000);
        meer_drv_set_freq(fd, 482);
        usleep(500000);
        meer_drv_set_freq(fd, 485);
        usleep(500000);
        meer_drv_set_freq(fd, 487);
        usleep(500000);
        meer_drv_set_freq(fd, 490);
        usleep(500000);
        meer_drv_set_freq(fd, 492);
        usleep(500000);
        meer_drv_set_freq(fd, 495);
        usleep(500000);
        meer_drv_set_freq(fd, 496);
        usleep(500000);
        meer_drv_set_freq(fd, 500);
        usleep(500000);
        meer_drv_set_freq(fd, 503);
        usleep(500000);
        meer_drv_set_freq(fd, 506);
        usleep(500000);
        meer_drv_set_freq(fd, 509);
        usleep(500000);
        meer_drv_set_freq(fd, 512);
        usleep(500000);
        meer_drv_set_freq(fd, 515);
        usleep(500000);
        meer_drv_set_freq(fd, 518);
        usleep(500000);
        meer_drv_set_freq(fd, 521);
        usleep(500000);
        meer_drv_set_freq(fd, 525);
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

	uart_read_register(fd, 0x01, 0x00);
	uart_read_register(fd, 0x01, 0x57);
	uart_read_register(fd, 0x01, 0x58);
	uart_read_register(fd, 0x01, 0x59);

	return fd;
}

//测试主程序
void set_work(int fd,uint8_t* header,int pheader_len,uint8_t* target,int chipId)
{
	struct work work_temp;
	memcpy(work_temp.target, target, 32); // 难度目标配置
	memcpy(work_temp.header, header, pheader_len); // meer区块头
	meer_drv_set_work(fd, &work_temp, chipId); // 对算力板下任务
}

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "uart.h"
#include "meer_drv.h"
#include "meer.h"

extern int init_drv(int num_of_chips,char *path);
extern void set_work(int fd,uint8_t* header,int pheader_len,uint8_t* target,int chipId);
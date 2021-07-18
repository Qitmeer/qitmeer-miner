#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char* argv[])
{
    uint64_t max = 0xffffffffffffffff;
    int num = 14;
    int group = 3;
    uint64_t unit = max/group/num;
    uint64_t index = 0;
    for(int i=0;i<num;i++){
        for(int j=0;j<group;j++){
            uint64_t step = unit * index;
            index++;
            printf("\nindex %llu chips %d group %d nonce %llu \n",index,i,j,step);
        }
    }
}
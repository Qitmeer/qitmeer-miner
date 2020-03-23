void search_circle(const unsigned int *arg_value,unsigned long length,unsigned int *ret);
void stop_solver(void *ppDetectInfo);
void init_solver(int deviceID,void **ppDetectInfo,int expand,int ntrims,int genablocks,int genatpb,int genbtpb,int trimtpb,int tailtpb,int recoverblocks,int recovertpb);
int run_solver(int deviceID,void *ppDetectInfo,char* header,int headerLen,unsigned int offset,unsigned int range,unsigned char* target,unsigned int *Nonce,unsigned int *CycleNonces,double *average);
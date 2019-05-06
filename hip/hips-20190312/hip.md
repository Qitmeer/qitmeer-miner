<pre>
  BIP: 1
  Layer: Block
  Title: Convert Block header time from uint32 to uint64 
  Author: james
  Comments-Summary: No comments yet.
  Comments-URI: 
  Status: Draft
  Type: Standards 
  Created: 2019-03-12
</pre>

==Abstract==

This hIP describes a new type of HLC block header time .Now the block length is 124 and the type of time is uint32
    
    core/types/block.go
    #20 const MaxBlockHeaderPayload = 4 + (hash.HashSize * 3) + 4 + 8 + 4 + 8
    #115 // TODO fix time ambiguous
    	sec := uint32(bh.Timestamp.Unix())
    #104 // TODO fix time ambiguous
        return s.ReadElements(r, &bh.Version, &bh.ParentRoot, &bh.TxRoot,
         		&bh.StateRoot, &bh.Difficulty, &bh.Height, (*s.Uint32Time)(&bh.Timestamp),
         		&bh.Nonce)

Now this hip fix this 
    
         core/types/block.go
            #20 const MaxBlockHeaderPayload = 4 + (hash.HashSize * 3) + 4 + 8 + 8 + 8
            #115 // TODO fix time ambiguous
            	sec := uint64(bh.Timestamp.Unix())
            #104 // TODO fix time ambiguous
                return s.ReadElements(r, &bh.Version, &bh.ParentRoot, &bh.TxRoot,
                 		&bh.StateRoot, &bh.Difficulty, &bh.Height, (*s.Int64Time)(&bh.Timestamp),
                 		&bh.Nonce)
==Motivation==

Because the opencl miner of the algorithm of blake2b needs the length of (header data)is Multiple of 8, if do this , the opencl has the best performance.
    
        ulong m[16] = {	headerIn[0], headerIn[1],
        	                headerIn[2], headerIn[3],
        	                headerIn[4], headerIn[5],
        	                headerIn[6], headerIn[7],
        	                headerIn[8], headerIn[9], headerIn[10], headerIn[11], headerIn[12], headerIn[13], headerIn[14], nonce };
        	                
The other reason the original author want the type of time is int64. 
    
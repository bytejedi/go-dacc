package types

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/daccproject/go-dacc/common"
	"github.com/daccproject/go-dacc/crypto/sha3"
	"github.com/daccproject/go-dacc/ethdb"
	"github.com/daccproject/go-dacc/log"
	"github.com/daccproject/go-dacc/rlp"
	"github.com/daccproject/go-dacc/trie"
)

type DposContext struct {
	epochTrie     *trie.Trie // 记录每个周期的验证人,就一条数据，k: []byte("validator") v: 经过rlp处理后的validators列表
	delegateTrie  *trie.Trie // 记录验证人以及对应投票人，k: append(candidate, delegator...) v: delegator
	voteTrie      *trie.Trie // 记录投票人对应验证人, k: delegator v: candidate
	candidateTrie *trie.Trie // 记录候选人, k: candidate v: candidate
	mintCntTrie   *trie.Trie // 记录验证人在周期内的出块数, k: append(第几个周期, validator.Bytes()...) v: 出块数

	db ethdb.Database
}

var (
	epochPrefix     = []byte("epoch-")
	delegatePrefix  = []byte("delegate-")
	votePrefix      = []byte("vote-")
	candidatePrefix = []byte("candidate-")
	mintCntPrefix   = []byte("mintCnt-")
)

func NewEpochTrie(root common.Hash, diskdb ethdb.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, epochPrefix, trie.NewDatabase(diskdb))
}

func NewDelegateTrie(root common.Hash, diskdb ethdb.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, delegatePrefix, trie.NewDatabase(diskdb))
}

func NewVoteTrie(root common.Hash, diskdb ethdb.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, votePrefix, trie.NewDatabase(diskdb))
}

func NewCandidateTrie(root common.Hash, diskdb ethdb.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, candidatePrefix, trie.NewDatabase(diskdb))
}

func NewMintCntTrie(root common.Hash, diskdb ethdb.Database) (*trie.Trie, error) {
	return trie.NewTrieWithPrefix(root, mintCntPrefix, trie.NewDatabase(diskdb))
}

func NewDposContext(db ethdb.Database) (*DposContext, error) {
	epochTrie, err := NewEpochTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	delegateTrie, err := NewDelegateTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	voteTrie, err := NewVoteTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	candidateTrie, err := NewCandidateTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	mintCntTrie, err := NewMintCntTrie(common.Hash{}, db)
	if err != nil {
		return nil, err
	}
	return &DposContext{
		epochTrie:     epochTrie,
		delegateTrie:  delegateTrie,
		voteTrie:      voteTrie,
		candidateTrie: candidateTrie,
		mintCntTrie:   mintCntTrie,
		db:            db,
	}, nil
}

func NewDposContextFromProto(db ethdb.Database, ctxProto *DposContextProto) (*DposContext, error) {
	epochTrie, err := NewEpochTrie(ctxProto.EpochHash, db)
	if err != nil {
		log.Info("error epochTrie:" + err.Error())
		return nil, err
	}
	delegateTrie, err := NewDelegateTrie(ctxProto.DelegateHash, db)
	if err != nil {
		log.Info("error delegateTrie:" + err.Error())
		return nil, err
	}
	voteTrie, err := NewVoteTrie(ctxProto.VoteHash, db)
	if err != nil {
		log.Info("error voteTrie:" + err.Error())
		return nil, err
	}
	candidateTrie, err := NewCandidateTrie(ctxProto.CandidateHash, db)
	if err != nil {
		log.Info("error candidateTrie:" + err.Error())
		return nil, err
	}
	mintCntTrie, err := NewMintCntTrie(ctxProto.MintCntHash, db)
	if err != nil {
		return nil, err
	}
	return &DposContext{
		epochTrie:     epochTrie,
		delegateTrie:  delegateTrie,
		voteTrie:      voteTrie,
		candidateTrie: candidateTrie,
		mintCntTrie:   mintCntTrie,
		db:            db,
	}, nil
}

func (d *DposContext) Copy() *DposContext {
	epochTrie := *d.epochTrie
	delegateTrie := *d.delegateTrie
	voteTrie := *d.voteTrie
	candidateTrie := *d.candidateTrie
	mintCntTrie := *d.mintCntTrie
	return &DposContext{
		epochTrie:     &epochTrie,
		delegateTrie:  &delegateTrie,
		voteTrie:      &voteTrie,
		candidateTrie: &candidateTrie,
		mintCntTrie:   &mintCntTrie,
		db:            d.db,
	}
}

func (d *DposContext) Root() (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, d.epochTrie.Hash())
	rlp.Encode(hw, d.delegateTrie.Hash())
	rlp.Encode(hw, d.candidateTrie.Hash())
	rlp.Encode(hw, d.voteTrie.Hash())
	rlp.Encode(hw, d.mintCntTrie.Hash())
	hw.Sum(h[:0])
	return h
}

func (d *DposContext) Snapshot() *DposContext {
	return d.Copy()
}

func (d *DposContext) RevertToSnapShot(snapshot *DposContext) {
	d.epochTrie = snapshot.epochTrie
	d.delegateTrie = snapshot.delegateTrie
	d.candidateTrie = snapshot.candidateTrie
	d.voteTrie = snapshot.voteTrie
	d.mintCntTrie = snapshot.mintCntTrie
}

func (d *DposContext) FromProto(dcp *DposContextProto) error {
	var err error
	d.epochTrie, err = NewEpochTrie(dcp.EpochHash, d.db)
	if err != nil {
		return err
	}
	d.delegateTrie, err = NewDelegateTrie(dcp.DelegateHash, d.db)
	if err != nil {
		return err
	}
	d.candidateTrie, err = NewCandidateTrie(dcp.CandidateHash, d.db)
	if err != nil {
		return err
	}
	d.voteTrie, err = NewVoteTrie(dcp.VoteHash, d.db)
	if err != nil {
		return err
	}
	d.mintCntTrie, err = NewMintCntTrie(dcp.MintCntHash, d.db)
	return err
}

type DposContextProto struct {
	EpochHash     common.Hash `json:"epochRoot"        gencodec:"required"`
	DelegateHash  common.Hash `json:"delegateRoot"     gencodec:"required"`
	CandidateHash common.Hash `json:"candidateRoot"    gencodec:"required"`
	VoteHash      common.Hash `json:"voteRoot"         gencodec:"required"`
	MintCntHash   common.Hash `json:"mintCntRoot"      gencodec:"required"`
}

func (d *DposContext) ToProto() *DposContextProto {
	return &DposContextProto{
		EpochHash:     d.epochTrie.Hash(),
		DelegateHash:  d.delegateTrie.Hash(),
		CandidateHash: d.candidateTrie.Hash(),
		VoteHash:      d.voteTrie.Hash(),
		MintCntHash:   d.mintCntTrie.Hash(),
	}
}

func (p *DposContextProto) Root() (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, p.EpochHash)
	rlp.Encode(hw, p.DelegateHash)
	rlp.Encode(hw, p.CandidateHash)
	rlp.Encode(hw, p.VoteHash)
	rlp.Encode(hw, p.MintCntHash)
	hw.Sum(h[:0])
	return h
}

func (d *DposContext) KickoutCandidate(candidateAddr common.Address) error {
	candidate := candidateAddr.Bytes()
	err := d.candidateTrie.TryDelete(candidate)
	if err != nil {
		if _, ok := err.(*trie.MissingNodeError); !ok {
			return err
		}
	}
	iter := trie.NewIterator(d.delegateTrie.PrefixIterator(candidate))
	for iter.Next() {
		delegator := iter.Value
		key := append(candidate, delegator...)
		err = d.delegateTrie.TryDelete(key)
		if err != nil {
			if _, ok := err.(*trie.MissingNodeError); !ok {
				return err
			}
		}
		v, err := d.voteTrie.TryGet(delegator)
		if err != nil {
			if _, ok := err.(*trie.MissingNodeError); !ok {
				return err
			}
		}
		if err == nil && bytes.Equal(v, candidate) {
			err = d.voteTrie.TryDelete(delegator)
			if err != nil {
				if _, ok := err.(*trie.MissingNodeError); !ok {
					return err
				}
			}
		}
	}
	return nil
}

func (d *DposContext) BecomeCandidate(candidateAddr common.Address) error {
	candidate := candidateAddr.Bytes()
	return d.candidateTrie.TryUpdate(candidate, candidate)
}

func (d *DposContext) Delegate(delegatorAddr, candidateAddr common.Address) error {
	delegator, candidate := delegatorAddr.Bytes(), candidateAddr.Bytes()

	// the candidate must be candidate
	candidateInTrie, err := d.candidateTrie.TryGet(candidate)
	if err != nil {
		return err
	}
	if candidateInTrie == nil {
		return errors.New("invalid candidate to delegate")
	}

	// delete old candidate if exists
	oldCandidate, err := d.voteTrie.TryGet(delegator)
	if err != nil {
		if _, ok := err.(*trie.MissingNodeError); !ok {
			return err
		}
	}
	if oldCandidate != nil {
		d.delegateTrie.Delete(append(oldCandidate, delegator...))
	}
	if err = d.delegateTrie.TryUpdate(append(candidate, delegator...), delegator); err != nil {
		return err
	}
	return d.voteTrie.TryUpdate(delegator, candidate)
}

func (d *DposContext) UnDelegate(delegatorAddr, candidateAddr common.Address) error {
	delegator, candidate := delegatorAddr.Bytes(), candidateAddr.Bytes()

	// the candidate must be candidate
	candidateInTrie, err := d.candidateTrie.TryGet(candidate)
	if err != nil {
		return err
	}
	if candidateInTrie == nil {
		return errors.New("invalid candidate to undelegate")
	}

	oldCandidate, err := d.voteTrie.TryGet(delegator)
	if err != nil {
		return err
	}
	if !bytes.Equal(candidate, oldCandidate) {
		return errors.New("mismatch candidate to undelegate")
	}

	if err = d.delegateTrie.TryDelete(append(candidate, delegator...)); err != nil {
		return err
	}
	return d.voteTrie.TryDelete(delegator)
}

func (d *DposContext) CommitTo() (*DposContextProto, error) {
	epochRoot, err := d.epochTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	if err := d.epochTrie.TrieDB().Commit(epochRoot, false); err != nil {
		return nil, err
	}
	delegateRoot, err := d.delegateTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	if err := d.delegateTrie.TrieDB().Commit(delegateRoot, false); err != nil {
		return nil, err
	}
	voteRoot, err := d.voteTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	if err := d.voteTrie.TrieDB().Commit(voteRoot, false); err != nil {
		return nil, err
	}
	candidateRoot, err := d.candidateTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	if err := d.candidateTrie.TrieDB().Commit(candidateRoot, false); err != nil {
		return nil, err
	}
	mintCntRoot, err := d.mintCntTrie.Commit(nil)
	if err != nil {
		return nil, err
	}
	if err := d.mintCntTrie.TrieDB().Commit(mintCntRoot, false); err != nil {
		return nil, err
	}
	/*
		epochRoot, err := d.epochTrie.Commit(nil)
		if err != nil {
			return nil, err
		}
		delegateRoot, err := d.delegateTrie.Commit(nil)
		if err != nil {
			return nil, err
		}
		voteRoot, err := d.voteTrie.Commit(nil)
		if err != nil {
			return nil, err
		}
		candidateRoot, err := d.candidateTrie.Commit(nil)
		if err != nil {
			return nil, err
		}
		mintCntRoot, err := d.mintCntTrie.Commit(nil)
		if err != nil {
			return nil, err
		}*/
	return &DposContextProto{
		EpochHash:     epochRoot,
		DelegateHash:  delegateRoot,
		VoteHash:      voteRoot,
		CandidateHash: candidateRoot,
		MintCntHash:   mintCntRoot,
	}, nil
}

func (d *DposContext) CandidateTrie() *trie.Trie          { return d.candidateTrie }
func (d *DposContext) DelegateTrie() *trie.Trie           { return d.delegateTrie }
func (d *DposContext) VoteTrie() *trie.Trie               { return d.voteTrie }
func (d *DposContext) EpochTrie() *trie.Trie              { return d.epochTrie }
func (d *DposContext) MintCntTrie() *trie.Trie            { return d.mintCntTrie }
func (d *DposContext) DB() ethdb.Database                 { return d.db }
func (dc *DposContext) SetEpoch(epoch *trie.Trie)         { dc.epochTrie = epoch }
func (dc *DposContext) SetDelegate(delegate *trie.Trie)   { dc.delegateTrie = delegate }
func (dc *DposContext) SetVote(vote *trie.Trie)           { dc.voteTrie = vote }
func (dc *DposContext) SetCandidate(candidate *trie.Trie) { dc.candidateTrie = candidate }
func (dc *DposContext) SetMintCnt(mintCnt *trie.Trie)     { dc.mintCntTrie = mintCnt }

// GetValidators 通过DposContext返回epochTrie中的validators
func (dc *DposContext) GetValidators() ([]common.Address, error) {
	var validators []common.Address
	key := []byte("validator")
	validatorsRLP := dc.epochTrie.Get(key)
	if err := rlp.DecodeBytes(validatorsRLP, &validators); err != nil {
		return nil, fmt.Errorf("failed to decode validators: %s", err)
	}
	return validators, nil
}

func (dc *DposContext) SetValidators(validators []common.Address) error {
	key := []byte("validator")
	validatorsRLP, err := rlp.EncodeToBytes(validators)
	if err != nil {
		return fmt.Errorf("failed to encode validators to rlp bytes: %s", err)
	}
	dc.epochTrie.Update(key, validatorsRLP)
	return nil
}

package privval

import (
	"fmt"

	"github.com/lazyledger/lazyledger-core/crypto"
	cryptoenc "github.com/lazyledger/lazyledger-core/crypto/encoding"
	privvalproto "github.com/lazyledger/lazyledger-core/proto/tendermint/privval"
	"github.com/lazyledger/lazyledger-core/types"
)

func DefaultValidationRequestHandler(
	privVal types.PrivValidator,
	req privvalproto.Message,
	chainID string,
) (privvalproto.Message, error) {
	var (
		res privvalproto.Message
		err error
	)

	switch r := req.Sum.(type) {
	case *privvalproto.Message_PubKeyRequest:
		if r.PubKeyRequest.GetChainId() != chainID {
			res = mustWrapMsg(&privvalproto.SignedVoteResponse{
				Vote: nil, Error: &privvalproto.RemoteSignerError{
					Code: 0, Description: "unable to provide pubkey"}})
			return res, fmt.Errorf("want chainID: %s, got chainID: %s", r.PubKeyRequest.GetChainId(), chainID)
		}

		var pubKey crypto.PubKey
		pubKey, err = privVal.GetPubKey()
		pk, err := cryptoenc.PubKeyToProto(pubKey)
		if err != nil {
			return res, err
		}

		if err != nil {
			res = mustWrapMsg(&privvalproto.PubKeyResponse{
				PubKey: nil, Error: &privvalproto.RemoteSignerError{Code: 0, Description: err.Error()}})
		} else {
			res = mustWrapMsg(&privvalproto.PubKeyResponse{PubKey: &pk, Error: nil})
		}

	case *privvalproto.Message_SignVoteRequest:
		if r.SignVoteRequest.ChainId != chainID {
			res = mustWrapMsg(&privvalproto.SignedVoteResponse{
				Vote: nil, Error: &privvalproto.RemoteSignerError{
					Code: 0, Description: "unable to sign vote"}})
			return res, fmt.Errorf("want chainID: %s, got chainID: %s", r.SignVoteRequest.GetChainId(), chainID)
		}

		vote := r.SignVoteRequest.Vote

		err = privVal.SignVote(chainID, vote)
		if err != nil {
			res = mustWrapMsg(&privvalproto.SignedVoteResponse{
				Vote: nil, Error: &privvalproto.RemoteSignerError{Code: 0, Description: err.Error()}})
		} else {
			res = mustWrapMsg(&privvalproto.SignedVoteResponse{Vote: vote, Error: nil})
		}

	case *privvalproto.Message_SignProposalRequest:
		if r.SignProposalRequest.GetChainId() != chainID {
			res = mustWrapMsg(&privvalproto.SignedVoteResponse{
				Vote: nil, Error: &privvalproto.RemoteSignerError{
					Code:        0,
					Description: "unable to sign proposal"}})
			return res, fmt.Errorf("want chainID: %s, got chainID: %s", r.SignProposalRequest.GetChainId(), chainID)
		}

		proposal := r.SignProposalRequest.Proposal

		err = privVal.SignProposal(chainID, proposal)
		if err != nil {
			res = mustWrapMsg(&privvalproto.SignedProposalResponse{
				Proposal: nil, Error: &privvalproto.RemoteSignerError{Code: 0, Description: err.Error()}})
		} else {
			res = mustWrapMsg(&privvalproto.SignedProposalResponse{Proposal: proposal, Error: nil})
		}
	case *privvalproto.Message_PingRequest:
		err, res = nil, mustWrapMsg(&privvalproto.PingResponse{})

	default:
		err = fmt.Errorf("unknown msg: %v", r)
	}

	return res, err
}

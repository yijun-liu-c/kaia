package vm

import (
	"encoding/json"
	"github.com/kaiachain/kaia/common"
	"github.com/kaiachain/kaia/common/hexutil"
	"math/big"
	"sync/atomic"
)

type stateMap = map[common.Address]*account

type account struct {
	Balance *big.Int                    `json:"balance,omitempty"`
	Code    []byte                      `json:"code,omitempty"`
	Nonce   uint64                      `json:"nonce,omitempty"`
	Storage map[common.Hash]common.Hash `json:"storage,omitempty"`
	empty   bool
}

func (a *account) exists() bool {
	return a.Nonce > 0 || len(a.Code) > 0 || len(a.Storage) > 0 || (a.Balance != nil && a.Balance.Sign() != 0)
}

type accountMarshaling struct {
	Balance *hexutil.Big
	Code    hexutil.Bytes
}

type PrestateTracer struct {
	env       *EVM
	pre       stateMap
	post      stateMap
	to        common.Address
	config    prestateTracerConfig
	interrupt atomic.Bool // Atomic flag to signal execution interruption
	reason    error       // Textual reason for the interruption
	created   map[common.Address]bool
	deleted   map[common.Address]bool
}

type prestateTracerConfig struct {
	DiffMode bool `json:"diffMode"` // If true, this tracer will return state modifications
}

func NewPrestateTracer() Tracer {
	pt := &PrestateTracer{
		pre:       make(stateMap),
		post:      make(stateMap),
		to:        common.Address{},
		config:    prestateTracerConfig{},
		interrupt: atomic.Bool{},
		reason:    nil,
		created:   make(map[common.Address]bool),
		deleted:   make(map[common.Address]bool),
	}
	return pt
}

// Transaction start
func (pt *PrestateTracer) CaptureTxStart(uint64) {}
func (pt *PrestateTracer) CaptureTxEnd(uint64)   {}
func (pt *PrestateTracer) CaptureStart(env *EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
}
func (pt *PrestateTracer) CaptureEnd(output []byte, gasUsed uint64, err error) {}
func (pt *PrestateTracer) CaptureEnter(typ OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
}
func (pt *PrestateTracer) CaptureExit(output []byte, gasUsed uint64, err error) {}
func (pt *PrestateTracer) CaptureState(env *EVM, pc uint64, op OpCode, gas, cost, ccLeft, ccOpcode uint64, scope *ScopeContext, depth int, err error) {
	if err != nil {
		return
	}
	stackData := scope.Stack.Data()
	_ = len(stackData)
	_ = scope.Contract.Address()
}

func (pt *PrestateTracer) CaptureFault(env *EVM, pc uint64, op OpCode, gas, cost, ccLeft, ccOpcode uint64, scope *ScopeContext, depth int, err error) {
}

func (pt *PrestateTracer) GetResult() (json.RawMessage, error) {
	res := []byte(`"success"`)
	return json.RawMessage(res), nil
}

func (pt *PrestateTracer) Stop(err error) {
	pt.interrupt.Store(true)
}

package localstatequery

import (
	"fmt"

	"github.com/cloudstruct/go-ouroboros-network/protocol"
)

// Server implements the LocalStateQuery server
type Server struct {
	*protocol.Protocol
	config                        *Config
	enableGetChainBlockNo         bool
	enableGetChainPoint           bool
	enableGetRewardInfoPoolsBlock bool
}

// NewServer returns a new LocalStateQuery server object
func NewServer(protoOptions protocol.ProtocolOptions, cfg *Config) *Server {
	s := &Server{
		config: cfg,
	}
	protoConfig := protocol.ProtocolConfig{
		Name:                protocolName,
		ProtocolId:          protocolId,
		Muxer:               protoOptions.Muxer,
		ErrorChan:           protoOptions.ErrorChan,
		Mode:                protoOptions.Mode,
		Role:                protocol.ProtocolRoleServer,
		MessageHandlerFunc:  s.messageHandler,
		MessageFromCborFunc: NewMsgFromCbor,
		StateMap:            StateMap,
		InitialState:        stateIdle,
	}
	// Enable version-dependent features
	if protoOptions.Version >= 10 {
		s.enableGetChainBlockNo = true
		s.enableGetChainPoint = true
	}
	if protoOptions.Version >= 11 {
		s.enableGetRewardInfoPoolsBlock = true
	}
	s.Protocol = protocol.New(protoConfig)
	return s
}

func (s *Server) messageHandler(msg protocol.Message, isResponse bool) error {
	var err error
	switch msg.Type() {
	case MessageTypeAcquire:
		err = s.handleAcquire(msg)
	case MessageTypeQuery:
		err = s.handleQuery(msg)
	case MessageTypeRelease:
		err = s.handleRelease()
	case MessageTypeReacquire:
		err = s.handleReAcquire(msg)
	case MessageTypeAcquireNoPoint:
		err = s.handleAcquire(msg)
	case MessageTypeReacquireNoPoint:
		err = s.handleReAcquire(msg)
	case MessageTypeDone:
		err = s.handleDone()
	default:
		err = fmt.Errorf("%s: received unexpected message type %d", protocolName, msg.Type())
	}
	return err
}

func (s *Server) handleAcquire(msg protocol.Message) error {
	if s.config.AcquireFunc == nil {
		return fmt.Errorf("received local-state-query Acquire message but no callback function is defined")
	}
	switch msgAcquire := msg.(type) {
	case *MsgAcquire:
		// Call the user callback function
		return s.config.AcquireFunc(msgAcquire.Point)
	case *MsgAcquireNoPoint:
		// Call the user callback function
		return s.config.AcquireFunc(nil)
	}
	return nil
}

func (s *Server) handleQuery(msg protocol.Message) error {
	if s.config.QueryFunc == nil {
		return fmt.Errorf("received local-state-query Query message but no callback function is defined")
	}
	msgQuery := msg.(*MsgQuery)
	// Call the user callback function
	return s.config.QueryFunc(msgQuery.Query)
}

func (s *Server) handleRelease() error {
	if s.config.ReleaseFunc == nil {
		return fmt.Errorf("received local-state-query Release message but no callback function is defined")
	}
	// Call the user callback function
	return s.config.ReleaseFunc()
}

func (s *Server) handleReAcquire(msg protocol.Message) error {
	if s.config.ReAcquireFunc == nil {
		return fmt.Errorf("received local-state-query ReAcquire message but no callback function is defined")
	}
	switch msgReAcquire := msg.(type) {
	case *MsgReAcquire:
		// Call the user callback function
		return s.config.ReAcquireFunc(msgReAcquire.Point)
	case *MsgReAcquireNoPoint:
		// Call the user callback function
		return s.config.ReAcquireFunc(nil)
	}
	return nil
}

func (s *Server) handleDone() error {
	if s.config.DoneFunc == nil {
		return fmt.Errorf("received local-state-query Done message but no callback function is defined")
	}
	// Call the user callback function
	return s.config.DoneFunc()
}

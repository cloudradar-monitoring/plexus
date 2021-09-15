package control

import "errors"

func (m *MeshCentral) CloseReason() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	pending, ok := m.pendingActions["close"]
	if !ok {
		return nil
	}
	cause, _ := pending.String("cause")
	if cause != "" {
		return errors.New(cause)
	}
	return nil
}

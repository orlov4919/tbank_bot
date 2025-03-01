package tgbot

// специальная константа, которая задает переход в следующее состояние, по любому сообщению

const (
	TextEvent EventType = "text"
)

type ID = int

type EventType = string

type StateType = string

type state struct {
	transitions map[EventType]StateType
}

type StateMachine struct {
	current map[ID]StateType
	initial StateType
	states  map[StateType]state
}

type Transition struct {
	Event EventType
	Dst   StateType
}

type Transitions []Transition

type StateDesc struct {
	Name        StateType
	Transitions Transitions
}

type States []StateDesc

func NewStateMachine(initial StateType, states States) (*StateMachine, error) {

	mStates := make(map[StateType]state)

	for _, s := range states {
		state := state{
			transitions: make(map[EventType]StateType),
		}

		for _, t := range s.Transitions {
			state.transitions[t.Event] = t.Dst
		}

		mStates[s.Name] = state
	}

	if _, ok := mStates[initial]; !ok {
		return nil, NewErrMachineCreationFailed()
	}

	machine := &StateMachine{
		current: make(map[ID]StateType),
		states:  mStates,
		initial: initial,
	}

	return machine, nil
}

func (m *StateMachine) Current(id ID) StateType {
	if m.current[id] == "" {
		return m.initial
	}
	return m.current[id]
}

func (m *StateMachine) getNextState(id ID, event EventType) (StateType, error) {
	current := m.Current(id)

	next, ok := m.states[current].transitions[event]

	if !ok {

		next, ok := m.states[current].transitions[TextEvent]

		if !ok {
			return "", NewErrEventDeclined()
		}

		return next, nil
	}

	return next, nil
}

func (m *StateMachine) Transition(id ID, event EventType) (StateType, error) {
	next, err := m.getNextState(id, event)

	if err != nil {
		return m.current[id], NewErrTransitionFailed(id, event)
	}

	m.current[id] = next
	return next, nil
}

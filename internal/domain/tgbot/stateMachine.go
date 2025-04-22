package tgbot

// специальная константа, которая задает переход в следующее состояние, по любому сообщению

const (
	TextEvent Event = "text"
)

type ID = int64
type Event = string
type State = string

type state struct {
	transitions map[Event]State
}

type StateMachine struct {
	current map[ID]State
	initial State
	states  map[State]state
}

type Transition struct {
	Event Event
	Dst   State
}

type Transitions []Transition

type StateDesc struct {
	Name        State
	Transitions Transitions
}

type States []StateDesc

func NewStateMachine(initial State, states States) (*StateMachine, error) {
	mStates := make(map[State]state)

	for _, s := range states {
		state := state{
			transitions: make(map[Event]State),
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
		current: make(map[ID]State),
		states:  mStates,
		initial: initial,
	}

	return machine, nil
}

func (m *StateMachine) Current(id ID) State {
	if m.current[id] == "" {
		return m.initial
	}

	return m.current[id]
}

func (m *StateMachine) getNextState(id ID, event Event) (State, error) {
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

func (m *StateMachine) Transition(id ID, event Event) (State, error) {
	next, err := m.getNextState(id, event)

	if err != nil {
		return m.current[id], NewErrTransitionFailed(id, event)
	}

	m.current[id] = next

	return next, nil
}

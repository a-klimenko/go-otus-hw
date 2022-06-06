package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	Count int
	First *ListItem
	Last  *ListItem
}

func (l list) Len() int {
	return l.Count
}

func (l list) Front() *ListItem {
	return l.First
}

func (l list) Back() *ListItem {
	return l.Last
}

func (l *list) PushFront(v interface{}) *ListItem {
	l.Count++
	newNode := &ListItem{Value: v}

	if l.First != nil {
		newNode.Next = l.First
		l.First.Prev = newNode
	}

	if l.Last == nil {
		l.Last = newNode
	}

	l.First = newNode
	return newNode
}

func (l *list) PushBack(v interface{}) *ListItem {
	l.Count++
	newNode := &ListItem{Value: v}

	if l.Last != nil {
		newNode.Prev = l.Last
		l.Last.Next = newNode
	}

	if l.First == nil {
		l.First = newNode
	}

	l.Last = newNode
	return newNode
}

func (l *list) Remove(i *ListItem) {
	l.Count--

	if i.Prev == nil {
		l.First = i.Next
	} else {
		i.Prev.Next = i.Next
	}

	if i.Next == nil {
		l.Last = i.Prev
	} else {
		i.Next.Prev = i.Prev
	}
}

func (l *list) MoveToFront(i *ListItem) {
	if i.Prev == nil {
		return
	}

	if i.Next == nil {
		l.Last = i.Prev
	} else {
		i.Next.Prev = i.Prev
	}

	i.Prev.Next = i.Next
	i.Prev = nil
	i.Next = l.First
	l.First.Prev = i
	l.First = i
}

func NewList() List {
	return new(list)
}

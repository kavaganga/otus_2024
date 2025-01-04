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
	//List // Remove me after realization.
	// Place your code here.
	items map[*ListItem]any
	count int
	front *ListItem
	back  *ListItem
}

func NewList() List {
	return &list{items: make(map[*ListItem]any), count: 0, front: nil, back: nil}
}

func (l *list) Len() int {
	return l.count
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	item := &ListItem{Value: v, Next: l.front, Prev: nil}
	if l.front != nil {
		l.front.Prev = item
	}
	l.front = item
	l.items[item] = v

	l.count++

	if l.count == 1 {
		l.back = item
	}
	return item
}

func (l *list) PushBack(v interface{}) *ListItem {
	item := &ListItem{Value: v, Next: nil, Prev: nil}
	if l.back != nil {
		l.back.Next = item
		item.Prev = l.back
	}
	l.back = item
	l.items[item] = v
	l.count++

	if l.count == 1 {
		l.front = item
	}

	return item
}

func (l *list) Remove(i *ListItem) {
	if l.front == i {
		l.front = i.Next
	}

	if l.back == i {
		l.back = i.Prev
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev

	}
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}

	l.count--
	delete(l.items, i)
}

func (l *list) MoveToFront(i *ListItem) {
	if l.front == i {
		return
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev

	}
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}

	l.front.Prev = i
	i.Next = l.front
	i.Prev = nil
	l.front = i
}

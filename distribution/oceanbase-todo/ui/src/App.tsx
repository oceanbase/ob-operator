import 'bulma/css/bulma.min.css'
import React, { useEffect } from 'react'
import './App.css'

import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'

type TodoItem = {
  id: number
  title: string
  description: string
  finishedAt?: Date
  createdAt: Date
  updatedAt: Date
  deletedAt?: Date

  editing?: boolean
  deleting?: boolean
}

function App() {
  const [list, setList] = React.useState<TodoItem[]>([])
  const [editBuffer, setEditBuffer] = React.useState<TodoItem[]>([])
  const [title, setTitle] = React.useState('')
  const [description, setDescription] = React.useState('')
  const [submit, setSubmit] = React.useState(false)

  const query = useQuery({ queryKey: ['todos'], queryFn: () => fetch("/api/todos").then(res => res.json()) })
  useEffect(() => {
    if (query.data) {
      setList(query.data.todos)
      setEditBuffer(query.data.todos)
    }
  }, [query.data])

  const queryClient = useQueryClient()
  const creation = useMutation({
    mutationFn: (data: { title: string, description: string }) => fetch("/api/todos", {
      method: "PUT",
      body: JSON.stringify(data),
    }).then(res => res.json()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] })
      setTitle('')
      setDescription('')
      setSubmit(false)
    }
  })

  const update = useMutation({
    mutationFn: (data: { id: number, title: string, description: string }): Promise<TodoItem> => fetch(`/api/todos/${data.id}`, {
      method: "PATCH",
      body: JSON.stringify({
        title: data.title,
        description: data.description,
      })
    }).then(res => res.json()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] })
    }
  })

  const deletion = useMutation({
    mutationFn: (id: number) => fetch(`/api/todos/${id}`, { method: "DELETE" }).then(res => res.json()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] })
    }
  })

  const done = useMutation({
    mutationFn: (data: {
      id: number,
      done: boolean
    }): Promise<TodoItem> => fetch(`/api/todos/${data.id}`, {
      method: "PATCH", body: JSON.stringify({
        id: data.id,
        finishedAt: data.done ? new Date() : null,
        clearFinishedAt: !data.done,
      })
    }).then(res => res.json()),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] })
    }
  })

  return (
    <div>
      <div className="container">
        <div className="is-size-2 mb-6">OceanBase TODO List</div>
        {list.map((item, index) => (
          <div key={index} className="box columns mb-6 is-align-items-center">
            <div className="column is-three-quarters-tablet is-full-mobile" style={{ textAlign: "left" }}>
              {item.editing ? (
                <div style={{ flex: 1, width: "100%" }}>
                  <div className="field">
                    <p className="control">
                      <input className="input" placeholder="Title" value={editBuffer[index].title} onInput={(e) => {
                        setEditBuffer(editBuffer.map((m, i) => {
                          if (i === index) {
                            return { ...m, title: e.currentTarget.value }
                          }
                          return m
                        }))
                      }} />
                    </p>
                  </div>
                  <div className="field">
                    <p className="control">
                      <textarea className="textarea" rows={1} placeholder="Description" value={editBuffer[index].description} onInput={(e) => {
                        setEditBuffer(editBuffer.map((m, i) => {
                          if (i === index) {
                            return { ...m, description: e.currentTarget.value }
                          }
                          return m
                        }))
                      }} />
                    </p>
                  </div>
                </div>
              ) : (
                <>
                  <div className="is-size-5">{index + 1}. {item.title} {item.finishedAt && <span>✅</span>}</div>
                  <div dangerouslySetInnerHTML={{ __html: item.description }}>
                  </div>
                </>

              )}
            </div>
            {item.editing ? <>
              <div className="button column" onClick={() => {
                update.mutate({
                  id: item.id,
                  title: editBuffer[index].title,
                  description: editBuffer[index].description,
                })
                console.log('update', editBuffer[index])
              }}>Submit</div>
              <div className="button column" onClick={() => {
                setList(list.map((m, i) => {
                  if (i === index) {
                    return { ...m, editing: !m.editing }
                  }
                  return m
                }))
              }}>Cancel</div>
            </> :
              item.deleting ? <>
                <div className="button is-danger column" onClick={() => {
                  deletion.mutate(item.id)
                }}>Confirm</div>
                <div className="button column" onClick={() => {
                  setList(list.map((m, i) => {
                    if (i === index) {
                      return { ...m, deleting: !m.deleting }
                    }
                    return m
                  }))
                }}>Cancel</div>
              </> :
                <>
                  <div className="button column" onClick={() => {
                    setEditBuffer(editBuffer.map((m, i) => {
                      if (i === index) {
                        return { ...m, title: item.title, description: item.description }
                      }
                      return m
                    }))
                    setList(list.map((m, i) => {
                      if (i === index) {
                        return { ...m, editing: !m.editing }
                      }
                      return m
                    }))
                  }}>Edit</div>
                  {item.finishedAt ?
                    <div className="button column"
                      onClick={() => {
                        done.mutate({ id: item.id, done: false })
                      }}>Undone</div> :
                    <div className="button column"
                      onClick={() => {
                        done.mutate({ id: item.id, done: true })
                      }}>Done</div>
                  }
                  <div className="button is-danger column" onClick={() => {
                    setList(list.map((m, i) => {
                      if (i === index) {
                        return { ...m, deleting: !m.deleting }
                      }
                      return m
                    }))
                  }}>Delete</div>
                </>}
          </div>
        ))
        }
      </div >
      <div className="container box new-item">
        <div className="field">
          <label className="label" style={{ textAlign: "start" }}>New item</label>
          <p className="control">
            <input className={"input" + `${submit && !title ? ' is-danger' : ''}`} placeholder="Title" value={title} onInput={(e) => setTitle(e.currentTarget.value)} />
          </p>
          {submit && !title && <p className="help is-danger" style={{ textAlign: "start" }}>Title is required</p>}
        </div>
        <div className="field">
          <p className="control">
            <textarea className="textarea" placeholder="Description" value={description} onInput={(e) => setDescription(e.currentTarget.value)} />
          </p>
        </div>
        <div className="field has-text-left">
          <p className="control">
            <button className="button is-success" onClick={() => {
              setSubmit(true)
              if (!title) {
                return
              }
              creation.mutate({ title, description })
            }}>
              Submit
            </button>
          </p>
        </div>
      </div>
    </div >
  )
}

export default App

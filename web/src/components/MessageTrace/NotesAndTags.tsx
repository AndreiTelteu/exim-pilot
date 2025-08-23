import React, { useState, useEffect } from 'react';
import { apiService } from '../../services/api';

interface MessageNote {
  id: number;
  message_id: string;
  user_id: string;
  note: string;
  is_public: boolean;
  created_at: string;
  updated_at: string;
}

interface MessageTag {
  id: number;
  message_id: string;
  tag: string;
  color?: string;
  user_id: string;
  created_at: string;
}

interface NotesAndTagsProps {
  messageId: string;
}

export const NotesAndTags: React.FC<NotesAndTagsProps> = ({ messageId }) => {
  const [notes, setNotes] = useState<MessageNote[]>([]);
  const [tags, setTags] = useState<MessageTag[]>([]);
  const [popularTags, setPopularTags] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Form states
  const [newNote, setNewNote] = useState('');
  const [newNotePublic, setNewNotePublic] = useState(true);
  const [newTag, setNewTag] = useState('');
  const [newTagColor, setNewTagColor] = useState('#3b82f6');
  const [editingNote, setEditingNote] = useState<MessageNote | null>(null);
  
  // UI states
  const [showNoteForm, setShowNoteForm] = useState(false);
  const [showTagForm, setShowTagForm] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchNotesAndTags();
    fetchPopularTags();
  }, [messageId]);

  const fetchNotesAndTags = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [notesResponse, tagsResponse] = await Promise.all([
        apiService.get(`/messages/${messageId}/notes`),
        apiService.get(`/messages/${messageId}/tags`)
      ]);
      
      setNotes((notesResponse.data as any).notes || []);
      setTags((tagsResponse.data as any).tags || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load notes and tags');
    } finally {
      setLoading(false);
    }
  };

  const fetchPopularTags = async () => {
    try {
      const response = await apiService.get('/tags/popular?limit=10');
      setPopularTags((response.data as any).tags || []);
    } catch (err) {
      // Ignore errors for popular tags
    }
  };

  const handleCreateNote = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newNote.trim()) return;

    try {
      setSubmitting(true);
      const response = await apiService.post(`/messages/${messageId}/notes`, {
        note: newNote.trim(),
        is_public: newNotePublic
      });
      
      setNotes([response.data as MessageNote, ...notes]);
      setNewNote('');
      setShowNoteForm(false);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create note');
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdateNote = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingNote || !editingNote.note.trim()) return;

    try {
      setSubmitting(true);
      const response = await apiService.put(`/messages/${messageId}/notes/${editingNote.id}`, {
        note: editingNote.note.trim(),
        is_public: editingNote.is_public
      });
      
      setNotes(notes.map(note => note.id === editingNote.id ? response.data as MessageNote : note));
      setEditingNote(null);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update note');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteNote = async (noteId: number) => {
    if (!confirm('Are you sure you want to delete this note?')) return;

    try {
      await apiService.delete(`/messages/${messageId}/notes/${noteId}`);
      setNotes(notes.filter(note => note.id !== noteId));
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete note');
    }
  };

  const handleCreateTag = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTag.trim()) return;

    try {
      setSubmitting(true);
      const response = await apiService.post(`/messages/${messageId}/tags`, {
        tag: newTag.trim(),
        color: newTagColor
      });
      
      setTags([...tags, response.data as MessageTag]);
      setNewTag('');
      setShowTagForm(false);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create tag');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteTag = async (tagId: number) => {
    try {
      await apiService.delete(`/messages/${messageId}/tags/${tagId}`);
      setTags(tags.filter(tag => tag.id !== tagId));
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete tag');
    }
  };

  const handleQuickAddTag = async (tagName: string) => {
    try {
      const response = await apiService.post(`/messages/${messageId}/tags`, {
        tag: tagName,
        color: '#3b82f6'
      });
      
      setTags([...tags, response.data as MessageTag]);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to add tag');
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const getTagStyle = (color?: string) => {
    return {
      backgroundColor: color || '#3b82f6',
      color: 'white'
    };
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="text-sm text-red-700">{error}</div>
          <button
            onClick={() => setError(null)}
            className="mt-2 text-sm text-red-600 hover:text-red-800"
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Tags Section */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">Tags</h3>
          <button
            onClick={() => setShowTagForm(!showTagForm)}
            className="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded-md text-sm"
          >
            Add Tag
          </button>
        </div>

        {/* Tag Form */}
        {showTagForm && (
          <form onSubmit={handleCreateTag} className="mb-4 p-3 bg-gray-50 rounded-md">
            <div className="flex items-center space-x-3">
              <input
                type="text"
                value={newTag}
                onChange={(e) => setNewTag(e.target.value)}
                placeholder="Tag name"
                className="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm"
                required
              />
              <input
                type="color"
                value={newTagColor}
                onChange={(e) => setNewTagColor(e.target.value)}
                className="w-10 h-10 border border-gray-300 rounded-md"
              />
              <button
                type="submit"
                disabled={submitting}
                className="bg-green-600 hover:bg-green-700 text-white px-3 py-2 rounded-md text-sm disabled:opacity-50"
              >
                Add
              </button>
              <button
                type="button"
                onClick={() => setShowTagForm(false)}
                className="bg-gray-300 hover:bg-gray-400 text-gray-700 px-3 py-2 rounded-md text-sm"
              >
                Cancel
              </button>
            </div>
          </form>
        )}

        {/* Popular Tags */}
        {popularTags.length > 0 && (
          <div className="mb-4">
            <p className="text-sm text-gray-600 mb-2">Popular tags:</p>
            <div className="flex flex-wrap gap-2">
              {popularTags
                .filter(tag => !tags.some(t => t.tag === tag))
                .map(tag => (
                <button
                  key={tag}
                  onClick={() => handleQuickAddTag(tag)}
                  className="px-2 py-1 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded text-xs"
                >
                  + {tag}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Tags Display */}
        <div className="flex flex-wrap gap-2">
          {tags.map(tag => (
            <div
              key={tag.id}
              className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium"
              style={getTagStyle(tag.color)}
            >
              {tag.tag}
              <button
                onClick={() => handleDeleteTag(tag.id)}
                className="ml-2 text-white hover:text-gray-200"
              >
                ×
              </button>
            </div>
          ))}
          {tags.length === 0 && (
            <p className="text-sm text-gray-500">No tags added yet.</p>
          )}
        </div>
      </div>

      {/* Notes Section */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">Notes</h3>
          <button
            onClick={() => setShowNoteForm(!showNoteForm)}
            className="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded-md text-sm"
          >
            Add Note
          </button>
        </div>

        {/* Note Form */}
        {showNoteForm && (
          <form onSubmit={handleCreateNote} className="mb-4 p-3 bg-gray-50 rounded-md">
            <textarea
              value={newNote}
              onChange={(e) => setNewNote(e.target.value)}
              placeholder="Enter your note..."
              rows={3}
              className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
              required
            />
            <div className="flex items-center justify-between mt-3">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={newNotePublic}
                  onChange={(e) => setNewNotePublic(e.target.checked)}
                  className="mr-2"
                />
                <span className="text-sm text-gray-600">Public note</span>
              </label>
              <div className="space-x-2">
                <button
                  type="submit"
                  disabled={submitting}
                  className="bg-green-600 hover:bg-green-700 text-white px-3 py-2 rounded-md text-sm disabled:opacity-50"
                >
                  Save Note
                </button>
                <button
                  type="button"
                  onClick={() => setShowNoteForm(false)}
                  className="bg-gray-300 hover:bg-gray-400 text-gray-700 px-3 py-2 rounded-md text-sm"
                >
                  Cancel
                </button>
              </div>
            </div>
          </form>
        )}

        {/* Notes Display */}
        <div className="space-y-4">
          {notes.map(note => (
            <div key={note.id} className="border border-gray-200 rounded-md p-3">
              {editingNote?.id === note.id ? (
                <form onSubmit={handleUpdateNote}>
                  <textarea
                    value={editingNote.note}
                    onChange={(e) => setEditingNote({ ...editingNote, note: e.target.value })}
                    rows={3}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm mb-3"
                    required
                  />
                  <div className="flex items-center justify-between">
                    <label className="flex items-center">
                      <input
                        type="checkbox"
                        checked={editingNote.is_public}
                        onChange={(e) => setEditingNote({ ...editingNote, is_public: e.target.checked })}
                        className="mr-2"
                      />
                      <span className="text-sm text-gray-600">Public note</span>
                    </label>
                    <div className="space-x-2">
                      <button
                        type="submit"
                        disabled={submitting}
                        className="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded-md text-sm disabled:opacity-50"
                      >
                        Save
                      </button>
                      <button
                        type="button"
                        onClick={() => setEditingNote(null)}
                        className="bg-gray-300 hover:bg-gray-400 text-gray-700 px-3 py-1 rounded-md text-sm"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                </form>
              ) : (
                <>
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <p className="text-sm text-gray-900 whitespace-pre-wrap">{note.note}</p>
                      <div className="mt-2 flex items-center space-x-4 text-xs text-gray-500">
                        <span>By {note.user_id}</span>
                        <span>{formatTimestamp(note.created_at)}</span>
                        {!note.is_public && (
                          <span className="bg-yellow-100 text-yellow-800 px-2 py-0.5 rounded">Private</span>
                        )}
                        {note.created_at !== note.updated_at && (
                          <span>Updated {formatTimestamp(note.updated_at)}</span>
                        )}
                      </div>
                    </div>
                    <div className="flex space-x-2 ml-4">
                      <button
                        onClick={() => setEditingNote(note)}
                        className="text-blue-600 hover:text-blue-800 text-sm"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteNote(note.id)}
                        className="text-red-600 hover:text-red-800 text-sm"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                </>
              )}
            </div>
          ))}
          {notes.length === 0 && (
            <p className="text-sm text-gray-500">No notes added yet.</p>
          )}
        </div>
      </div>
    </div>
  );
};
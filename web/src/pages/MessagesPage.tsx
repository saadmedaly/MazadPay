import { useState, useEffect, useRef } from 'react'
import { 
  Search, 
  Send, 
  Paperclip, 
  MoreVertical, 
  User, 
  MessageSquare,
  Check,
  CheckCheck,
  Phone,
  Video,
  Info,
  AlertCircle,
  RefreshCw,
  Plus,
  X
} from 'lucide-react'
import { PageHeader } from '@/components/shared/PageHeader'
import { LoadingSpinner } from '@/components/shared/LoadingSpinner'
import { cn } from '@/lib/utils'
import { useConversations, useMessages, useSendMessage, useMarkAsRead, useCreateConversation } from '@/hooks/useMessages'
import type { Conversation } from '@/api/messages'
import { useUsers } from '@/hooks/useUsers'
import { formatRelative, formatDate } from '@/lib/formatters'
import { format } from 'date-fns'
import { ar } from 'date-fns/locale'
import { useAuthStore } from '@/stores/authStore'
import { useChatWebSocket } from '@/hooks/useChatWebSocket'

export function MessagesPage() {
  const { user: currentUser } = useAuthStore()
  const [selectedConvId, setSelectedConvId] = useState<string | null>(null)
  const [messageText, setMessageText] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [isNewChatModalOpen, setIsNewChatModalOpen] = useState(false)
  const [isGroupModalOpen, setIsGroupModalOpen] = useState(false)
  const [groupTitle, setGroupTitle] = useState('')
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([])
  const [userSearchQuery, setUserSearchQuery] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  // Real-time updates
  useChatWebSocket(selectedConvId)

  const { 
    data: conversationsData, 
    isLoading: isConvsLoading, 
    isError: isConvsError,
    refetch: refetchConvs 
  } = useConversations()
  const conversations = conversationsData ?? []

  const { 
    data: messagesData, 
    isLoading: isMsgsLoading,
    refetch: refetchMsgs
  } = useMessages(selectedConvId || '')
  const messages = messagesData ?? []

  const sendMessageMutation = useSendMessage()
  const markAsReadMutation = useMarkAsRead()
  const createConvMutation = useCreateConversation()

  const { data: usersData } = useUsers(userSearchQuery, 1)
  const users = usersData?.data ?? []
  const [lastCreatedConv, setLastCreatedConv] = useState<Conversation | null>(null)
  const selectedConv = conversations.find(c => c.id === selectedConvId) || (lastCreatedConv?.id === selectedConvId ? lastCreatedConv : null)

  const filteredConversations = conversations.filter(conv => {
    if (!searchQuery) return true
    const otherParticipant = conv.participants?.find(p => p.user_id !== currentUser?.id)?.user
    const name = conv.title || otherParticipant?.full_name || ''
    return name.toLowerCase().includes(searchQuery.toLowerCase())
  })

  // Scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  // Mark as read when conversation is selected
  useEffect(() => {
    if (selectedConvId && selectedConv?.unread_count && selectedConv.unread_count > 0) {
      markAsReadMutation.mutate(selectedConvId)
    }
  }, [selectedConvId, selectedConv?.unread_count])

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault()
    if (!messageText.trim() || !selectedConvId) return

    sendMessageMutation.mutate({
      conversationId: selectedConvId,
      payload: {
        type: 'text',
        content: messageText
      }
    }, {
      onSuccess: () => {
        setMessageText('')
      }
    })
  }


  const handleCreateChat = (userId: string) => {
    createConvMutation.mutate({
      type: 'direct',
      user_ids: [userId]
    }, {
      onSuccess: (newConv) => {
        setIsNewChatModalOpen(false)
        setUserSearchQuery('')
        setLastCreatedConv(newConv)
        setSelectedConvId(newConv.id)
      }
    })
  }


  if (isConvsLoading) return <LoadingSpinner fullPage label="جاري تحميل المحادثات..." />

  if (isConvsError) return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-surface-muted gap-4">
      <AlertCircle className="w-16 h-16 opacity-20 text-red-500" />
      <h2 className="text-white font-display font-bold text-xl">فشل في تحميل المحادثات</h2>
      <button 
        onClick={() => refetchConvs()} 
        className="mt-4 flex items-center gap-2 bg-mazad-primary text-white px-8 py-3 rounded-xl font-bold shadow-lg"
      >
        <RefreshCw className="w-4 h-4" />
        إعادة المحاولة
      </button>
    </div>
  )

  return (
    <div className="animate-fade-in flex flex-col h-[calc(100vh-140px)]" dir="rtl">
      <PageHeader 
        title="الرسائل"
        subtitle="تواصل مع المستخدمين وأجب على استفساراتهم"
      />

      <div className="flex-1 flex overflow-hidden bg-surface-card border border-surface-border rounded-2xl shadow-xl">
        {/* Conversations List */}
        <div className="w-80 border-l border-surface-border flex flex-col bg-surface-card/50">
          <div className="p-4 border-b border-surface-border">
            <div className="flex items-center gap-2 mb-3">
              <div className="flex-1 relative">
                <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
                <input 
                  type="text" 
                  placeholder="بحث عن محادثة..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full bg-surface-base border border-surface-border rounded-xl py-2 pr-10 pl-4 text-sm text-white focus:outline-none focus:ring-2 focus:ring-mazad-primary/50 transition-all"
                />
              </div>
                <button 
                  onClick={() => setIsNewChatModalOpen(true)}
                  className="p-2 bg-mazad-primary/10 text-mazad-primary hover:bg-mazad-primary hover:text-white rounded-xl transition-all shadow-sm"
                  title="محادثة جديدة"
                >
                  <Plus className="w-5 h-5" />
                </button>
                {/* New Group Button - visible to admins */}
                {currentUser?.role === 'admin' || currentUser?.role === 'super_admin' ? (
                  <button 
                    onClick={() => setIsGroupModalOpen(true)}
                    className="p-2 bg-mazad-success/10 text-mazad-success hover:bg-mazad-success hover:text-white rounded-xl transition-all shadow-sm ml-2"
                    title="إنشاء مجموعة"
                  >
                    <User className="w-5 h-5" />
                  </button>
                ) : null}
            </div>
          </div>

          <div className="flex-1 overflow-y-auto scrollbar-thin">
            {filteredConversations.length === 0 ? (
              <div className="p-10 text-center text-surface-muted">
                <MessageSquare className="w-10 h-10 mx-auto mb-2 opacity-20" />
                <p className="text-xs">
                  {searchQuery ? 'لا توجد نتائج للبحث' : 'لا توجد محادثات نشطة'}
                </p>
              </div>
            ) : filteredConversations.map((conv, idx) => {
              // For support/direct, find the other participant
              const otherParticipant = conv.participants?.find(p => p.user_id !== currentUser?.id)?.user
              const displayName = conv.title || otherParticipant?.full_name || 'مستخدم'
              const avatar = otherParticipant?.profile_pic_url

              return (
                <button
                  key={`conv-${conv.id}-${idx}`}
                  onClick={() => setSelectedConvId(conv.id)}
                  className={cn(
                    "w-full p-4 flex gap-3 items-start border-b border-surface-border/50 hover:bg-surface-border/30 transition-colors text-right",
                    selectedConvId === conv.id && "bg-mazad-primary/10 border-r-4 border-r-mazad-primary"
                  )}
                >
                  <div className="relative shrink-0">
                    <div className="w-12 h-12 rounded-full bg-surface-border flex items-center justify-center text-white font-bold text-lg overflow-hidden border border-surface-border">
                      {avatar ? (
                        <img src={avatar} alt={displayName} className="w-full h-full object-cover" />
                      ) : (
                        <User className="w-6 h-6 text-surface-muted" />
                      )}
                    </div>
                    {/* Simplified online status for now */}
                    {otherParticipant?.is_active && (
                      <div className="absolute bottom-0 left-0 w-3 h-3 rounded-full bg-mazad-success border-2 border-surface-card">
                        <div className="absolute inset-0 rounded-full bg-mazad-success animate-ping opacity-75" />
                      </div>
                    )}
                    {!otherParticipant?.is_active && (
                      <div className="absolute bottom-0 left-0 w-3 h-3 rounded-full bg-surface-muted border-2 border-surface-card" />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between items-center mb-1">
                      <h3 className="text-sm font-bold text-white truncate">{displayName}</h3>
                      <span className="text-[10px] text-surface-muted">
                        {conv.last_message_at ? formatRelative(conv.last_message_at) : ''}
                      </span>
                    </div>
                    <p className="text-xs text-surface-muted truncate leading-relaxed">
                      {conv.last_message_preview || 'لا توجد رسائل بعد'}
                    </p>
                  </div>
                  {conv.unread_count > 0 && (
                    <div className="bg-mazad-primary text-white text-[10px] font-bold min-w-[18px] h-[18px] rounded-full flex items-center justify-center px-1 shrink-0 mt-1">
                      {conv.unread_count}
                    </div>
                  )}
                </button>
              )
            })}
          </div>
        </div>

        {/* Chat Area */}
        <div className="flex-1 flex flex-col bg-surface-base/20">
          {selectedConvId ? (
            <>
              {/* Chat Header */}
              <div className="p-4 border-b border-surface-border flex justify-between items-center bg-surface-card/80 backdrop-blur-md">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-surface-border flex items-center justify-center overflow-hidden border border-surface-border">
                    {selectedConv?.participants?.find(p => p.user_id !== currentUser?.id)?.user?.profile_pic_url ? (
                      <img 
                        src={selectedConv.participants.find(p => p.user_id !== currentUser?.id)?.user?.profile_pic_url} 
                        className="w-full h-full object-cover"
                      />
                    ) : (
                      <User className="w-5 h-5 text-surface-muted" />
                    )}
                  </div>
                  <div>
                    <h3 className="text-sm font-bold text-white">
                      {selectedConv?.title || selectedConv?.participants?.find(p => p.user_id !== currentUser?.id)?.user?.full_name || 'محادثة'}
                    </h3>
                    <p className="text-[10px] text-mazad-success font-medium">
                      {selectedConv?.participants?.find(p => p.user_id !== currentUser?.id)?.user?.is_active ? 'متصل الآن' : 'غير متصل'}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button className="p-2 text-surface-muted hover:text-white hover:bg-surface-border rounded-lg transition-colors">
                    <Info className="w-4 h-4" />
                  </button>
                  <div className="w-px h-4 bg-surface-border mx-1" />
                  <button className="p-2 text-surface-muted hover:text-white hover:bg-surface-border rounded-lg transition-colors">
                    <MoreVertical className="w-4 h-4" />
                  </button>
                </div>
              </div>

              {/* Messages List */}
              <div className="flex-1 overflow-y-auto p-6 space-y-6 scrollbar-thin flex flex-col">
                {isMsgsLoading && messages.length === 0 ? (
                  <div className="flex-1 flex items-center justify-center">
                    <LoadingSpinner label="جاري تحميل الرسائل..." />
                  </div>
                ) : messages.length === 0 ? (
                  <div className="flex-1 flex flex-col items-center justify-center text-surface-muted opacity-50">
                    <MessageSquare className="w-12 h-12 mb-2" />
                    <p>لا توجد رسائل في هذه المحادثة</p>
                  </div>
                ) : (
                  <>
                    {messages.map((msg, idx) => {
                      const isMe = msg.sender?.id === currentUser?.id || msg.sender?.id === undefined
                      const showDate = idx === 0 || format(new Date(messages[idx-1].created_at), 'yyyy-MM-dd') !== format(new Date(msg.created_at), 'yyyy-MM-dd')

                      return (
                        <div key={`msg-${msg.id || idx}`} className="flex flex-col">
                          {showDate && (
                            <div className="mx-auto my-4">
                              <span className="text-[10px] font-bold text-surface-muted bg-surface-border/50 px-3 py-1 rounded-full uppercase tracking-widest">
                                {format(new Date(msg.created_at), 'dd MMMM yyyy', { locale: ar })}
                              </span>
                            </div>
                          )}
                          <div className={cn(
                            "flex flex-col max-w-[70%]",
                            isMe ? "mr-auto items-end" : "ml-auto items-start"
                          )}>
                            <div className={cn(
                              "p-4 rounded-2xl text-sm leading-relaxed shadow-sm",
                              isMe 
                                ? "bg-mazad-primary text-white rounded-tl-none" 
                                : "bg-surface-card border border-surface-border text-white rounded-tr-none"
                            )}>
                              {msg.content}
                              {msg.type === 'image' && msg.file_url && (
                                <img src={msg.file_url} className="rounded-lg mt-2 max-w-full" alt="Message" />
                              )}
                            </div>
                            <div className="flex items-center gap-1.5 mt-2 px-1">
                              <span className="text-[10px] text-surface-muted font-medium">
                                {format(new Date(msg.created_at), 'p', { locale: ar })}
                              </span>
                              {isMe && (
                                msg.status?.some(s => s.status === 'read') 
                                  ? <CheckCheck className="w-3 h-3 text-mazad-primary" /> 
                                  : <Check className="w-3 h-3 text-surface-muted" />
                              )}
                            </div>
                          </div>
                        </div>
                      )
                    })}
                    <div ref={messagesEndRef} />
                  </>
                )}
              </div>

              {/* Input Area */}
              <div className="p-4 bg-surface-card/80 backdrop-blur-md border-t border-surface-border">
                <form 
                  onSubmit={handleSend}
                  className="flex items-center gap-3"
                >
                  <button type="button" className="p-2 text-surface-muted hover:text-white hover:bg-surface-border rounded-lg transition-colors">
                    <Paperclip className="w-5 h-5" />
                  </button>
                  <div className="flex-1 relative">
                    <input 
                      type="text" 
                      value={messageText}
                      onChange={(e) => setMessageText(e.target.value)}
                      placeholder="اكتب رسالتك هنا..."
                      className="w-full bg-surface-base border border-surface-border rounded-xl py-3 px-4 text-sm text-white focus:outline-none focus:ring-2 focus:ring-mazad-primary/50 transition-all shadow-inner"
                      disabled={sendMessageMutation.isPending}
                    />
                  </div>
                  <button 
                    type="submit"
                    disabled={!messageText.trim() || sendMessageMutation.isPending}
                    className={cn(
                      "p-3 rounded-xl flex items-center justify-center transition-all shadow-lg",
                      messageText.trim() && !sendMessageMutation.isPending
                        ? "bg-mazad-primary text-white hover:bg-mazad-primary-dk shadow-mazad-primary/20 scale-100 active:scale-95" 
                        : "bg-surface-border text-surface-muted cursor-not-allowed"
                    )}
                  >
                    {sendMessageMutation.isPending ? <RefreshCw className="w-5 h-5 animate-spin" /> : <Send className="w-5 h-5" />}
                  </button>
                </form>
              </div>
            </>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center text-surface-muted gap-4">
              <div className="w-20 h-20 rounded-full bg-surface-border/30 flex items-center justify-center">
                <MessageSquare className="w-10 h-10 opacity-20" />
              </div>
              <div className="text-center">
                <h3 className="text-white font-bold">ابدأ محادثة</h3>
                <p className="text-sm">اختر مستخدماً من القائمة الجانبية لبدء المراسلة</p>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* New Group Modal */}
      {isGroupModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-surface-card border border-surface-border rounded-3xl w-full max-w-md shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
            <div className="p-6 border-b border-surface-border flex items-center justify-between">
              <h3 className="text-xl font-bold text-white">إنشاء مجموعة</h3>
              <button
                onClick={() => setIsGroupModalOpen(false)}
                className="p-2 hover:bg-surface-base rounded-full text-surface-muted transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <div className="p-4">
              <div className="mb-4">
                <input
                  type="text"
                  placeholder="عنوان المجموعة"
                  value={groupTitle}
                  onChange={e => setGroupTitle(e.target.value)}
                  className="w-full bg-surface-base border border-surface-border rounded-xl py-3 pr-4 pl-4 text-sm text-white focus:outline-none focus:ring-2 focus:ring-mazad-primary/50 transition-all"
                />
              </div>
              <div className="relative mb-4">
                <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
                <input
                  type="text"
                  placeholder="بحث عن مستخدم..."
                  value={userSearchQuery}
                  onChange={e => setUserSearchQuery(e.target.value)}
                  className="w-full bg-surface-base border border-surface-border rounded-xl py-3 pr-10 pl-4 text-sm text-white focus:outline-none focus:ring-2 focus:ring-mazad-primary/50 transition-all"
                />
              </div>
              <div className="max-h-60 overflow-y-auto scrollbar-thin space-y-2">
                {users.length === 0 ? (
                  <div className="py-10 text-center text-surface-muted">
                    <p className="text-sm">ابدأ الكتابة للبحث عن مستخدمين</p>
                  </div>
                ) : (
                  users.filter(u => u.id !== currentUser?.id).map(user => (
                    <button
                      key={user.id}
                      onClick={() => {
                        setSelectedUserIds(prev =>
                          prev.includes(user.id) ? prev.filter(id => id !== user.id) : [...prev, user.id]
                        );
                      }}
                      className={`group w-full flex items-center gap-3 p-3 rounded-xl transition-all text-right ${selectedUserIds.includes(user.id) ? 'bg-mazad-primary/20 border border-mazad-primary' : 'hover:bg-surface-base'}
                      `}
                    >
                      <div className="w-10 h-10 rounded-full bg-mazad-primary/10 flex items-center justify-center overflow-hidden border border-mazad-primary/20">
                        {user.profile_pic_url ? (
                          <img src={user.profile_pic_url} alt="" className="w-full h-full object-cover" />
                        ) : (
                          <User className="w-5 h-5 text-mazad-primary" />
                        )}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-bold text-white truncate">{user.full_name || 'مستخدم'}</div>
                        <div className="text-xs text-surface-muted truncate">{user.phone}</div>
                      </div>
                    </button>
                  ))
                )}
              </div>
              <div className="p-4 bg-surface-base/50 border-t border-surface-border flex justify-end">
                <button
                  onClick={() => setIsGroupModalOpen(false)}
                  className="mr-2 px-4 py-2 text-sm font-bold text-surface-muted hover:text-white transition-colors"
                >
                  إلغاء
                </button>
                <button
                  onClick={() => {
                    createConvMutation.mutate({
                      type: 'group',
                      user_ids: selectedUserIds,
                      title: groupTitle,
                    });
                    setIsGroupModalOpen(false);
                    setGroupTitle('');
                    setSelectedUserIds([]);
                  }}
                  disabled={selectedUserIds.length === 0 || !groupTitle.trim()}
                  className={`px-4 py-2 text-sm font-bold ${selectedUserIds.length === 0 || !groupTitle.trim() ? 'bg-surface-border text-surface-muted cursor-not-allowed' : 'bg-mazad-success text-white hover:bg-mazad-success-dk'} transition-colors`}
                >
                  إنشاء
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* New Chat Modal */}
      {isNewChatModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="bg-surface-card border border-surface-border rounded-3xl w-full max-w-md shadow-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
            <div className="p-6 border-b border-surface-border flex items-center justify-between">
              <h3 className="text-xl font-bold text-white">محادثة جديدة</h3>
              <button 
                onClick={() => setIsNewChatModalOpen(false)}
                className="p-2 hover:bg-surface-base rounded-full text-surface-muted transition-colors"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            
            <div className="p-4">
              <div className="relative mb-4">
                <Search className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-surface-muted" />
                <input 
                  type="text" 
                  autoFocus
                  placeholder="ابحث عن مستخدم بالاسم أو رقم الهاتف..."
                  value={userSearchQuery}
                  onChange={(e) => setUserSearchQuery(e.target.value)}
                  className="w-full bg-surface-base border border-surface-border rounded-xl py-3 pr-10 pl-4 text-sm text-white focus:outline-none focus:ring-2 focus:ring-mazad-primary/50 transition-all"
                />
              </div>

              <div className="max-h-60 overflow-y-auto scrollbar-thin space-y-2">
                {users.length === 0 ? (
                  <div className="py-10 text-center text-surface-muted">
                    <p className="text-sm">ابدأ الكتابة للبحث عن مستخدمين</p>
                  </div>
                ) : (
                  users.filter(u => u.id !== currentUser?.id).map((user, idx) => (
                    <button
                      key={`user-${user.id || idx}`}
                      onClick={() => handleCreateChat(user.id)}
                      className="group w-full flex items-center gap-3 p-3 rounded-xl hover:bg-surface-base border border-transparent hover:border-surface-border transition-all text-right"
                    >
                      <div className="w-10 h-10 rounded-full bg-mazad-primary/10 flex items-center justify-center overflow-hidden border border-mazad-primary/20">
                        {user.profile_pic_url ? (
                          <img src={user.profile_pic_url} alt="" className="w-full h-full object-cover" />
                        ) : (
                          <User className="w-5 h-5 text-mazad-primary" />
                        )}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="text-sm font-bold text-white truncate">{user.full_name || 'مستخدم بدون اسم'}</div>
                        <div className="text-xs text-surface-muted truncate">{user.phone}</div>
                      </div>
                      <Plus className="w-4 h-4 text-mazad-primary opacity-0 group-hover:opacity-100 transition-opacity" />
                    </button>
                  ))
                )}
              </div>
            </div>

            <div className="p-4 bg-surface-base/50 border-t border-surface-border">
              <button 
                onClick={() => setIsNewChatModalOpen(false)}
                className="w-full py-3 text-sm font-bold text-surface-muted hover:text-white transition-colors"
              >
                إلغاء
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

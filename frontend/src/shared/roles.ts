/** Проверка доступа по ролям (поддержка одного role и массива roles из API). */
export function hasTeacherAccess(user: { role?: string; roles?: string[] } | null): boolean {
  if (!user) return false
  const roles = user.roles ?? (user.role ? [user.role] : [])
  return roles.includes('teacher') || roles.includes('administrator')
}

export function hasAdminAccess(user: { role?: string; roles?: string[] } | null): boolean {
  if (!user) return false
  const roles = user.roles ?? (user.role ? [user.role] : [])
  return roles.includes('administrator')
}

/** Может ли пользователь редактировать модуль (владелец или администратор). */
export function canEditModule(
  user: { id: string; role?: string; roles?: string[] } | null,
  module: { owner_id?: string; is_owner?: boolean } | null
): boolean {
  if (!user || !module) return false
  if (module.is_owner) return true
  if (module.owner_id && module.owner_id === user.id) return true
  return hasAdminAccess(user)
}

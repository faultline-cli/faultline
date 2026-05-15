import { PrismaClient } from '@prisma/client'
const prisma = new PrismaClient()

export async function deleteUser(id: string) {
  prisma.user.delete({ where: { id } })
  return { success: true }
}

import { defineStore } from "pinia"

export const useSessionStore = defineStore("session", {
  state: () => ({
    token: localStorage.getItem("admin_token") || "",
    user: JSON.parse(localStorage.getItem("admin_user") || "null")
  }),
  actions: {
    save(payload) {
      this.token = payload.token
      this.user = payload.user
      localStorage.setItem("admin_token", payload.token)
      localStorage.setItem("admin_user", JSON.stringify(payload.user))
    },
    clear() {
      this.token = ""
      this.user = null
      localStorage.removeItem("admin_token")
      localStorage.removeItem("admin_user")
    }
  }
})

// Desktop sidebar collapse, remembered per browser.
const KEY = 'jevidocs-sidebar-collapsed'

function read(): boolean {
  try {
    return localStorage.getItem(KEY) === '1'
  } catch {
    return false
  }
}

class Sidebar {
  collapsed = $state(read())

  toggle(value = !this.collapsed) {
    this.collapsed = value
    try {
      localStorage.setItem(KEY, value ? '1' : '0')
    } catch {
      // storage unavailable: keep it for this page view only
    }
  }
}

export const sidebar = new Sidebar()

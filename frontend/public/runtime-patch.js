(() => {
  const yuanSymbol = '\uFFE5'
  const costDollarPlaceholder = '__KEEP_COST_DOLLAR__'
  const costCurrencyPattern = /((?:\u6210\u672C|cost)\s*[\uff1a:]\s*)[\uFFE5$](?=\s*\d)/gi
  const wideViewportMinWidth = 1920
  const dashboardWideSelector = '.mx-auto.w-full'
  const dashboardWideCardClass = 'max-w-[1380px]'

  const isDashboardWideCardElement = (element) => {
    const classList = element?.classList
    return (
      !!classList &&
      classList.contains('mx-auto') &&
      classList.contains('w-full') &&
      classList.contains(dashboardWideCardClass)
    )
  }

  const normalizeDollarText = (value) => {
    if (typeof value !== 'string') return value

    let next = value
    next = next.replace(costCurrencyPattern, `$1${costDollarPlaceholder}`)

    if (/^\s*\$\s*$/.test(next)) {
      next = next.replace('$', yuanSymbol)
    } else {
      next = next.replace(/\$(?=\s*\d)/g, yuanSymbol)
    }

    if (next.includes(costDollarPlaceholder)) {
      next = next.replaceAll(costDollarPlaceholder, '$')
    }

    return next
  }

  const patchTextNode = (node) => {
    const nextValue = normalizeDollarText(node.nodeValue)
    if (nextValue !== node.nodeValue) {
      node.nodeValue = nextValue
    }
  }

  const patchAttributes = (element) => {
    for (const name of ['placeholder', 'title', 'aria-label']) {
      if (!element.hasAttribute(name)) continue
      const current = element.getAttribute(name)
      const nextValue = normalizeDollarText(current)
      if (nextValue !== current) {
        element.setAttribute(name, nextValue)
      }
    }
  }

  const collectWideCardElements = (root) => {
    const elements = []
    if (window.location.pathname !== '/dashboard' || window.innerWidth < wideViewportMinWidth) {
      return elements
    }

    if (root?.nodeType !== Node.ELEMENT_NODE) return elements
    if (isDashboardWideCardElement(root)) {
      elements.push(root)
    }
    root.querySelectorAll?.(dashboardWideSelector).forEach((element) => {
      if (isDashboardWideCardElement(element)) {
        elements.push(element)
      }
    })
    return elements
  }

  const syncWideCardWidth = (root = document.body) => {
    const shouldExpand = window.location.pathname === '/dashboard' && window.innerWidth >= wideViewportMinWidth
    collectWideCardElements(root).forEach((element) => {
      element.style.maxWidth = shouldExpand ? 'none' : ''
    })
  }

  const patchTree = (root) => {
    if (!root) return
    if (root.nodeType === Node.TEXT_NODE) {
      patchTextNode(root)
      return
    }
    if (root.nodeType !== Node.ELEMENT_NODE) return
    patchAttributes(root)
    syncWideCardWidth(root)
    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
    while (walker.nextNode()) {
      patchTextNode(walker.currentNode)
    }
    root.querySelectorAll?.('*').forEach(patchAttributes)
  }

  const observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      if (mutation.type === 'characterData') {
        patchTextNode(mutation.target)
        continue
      }
      if (mutation.type === 'attributes') {
        patchAttributes(mutation.target)
        continue
      }
      mutation.addedNodes.forEach(patchTree)
    }
  })

  const start = () => {
    patchTree(document.body)
    syncWideCardWidth()
    window.addEventListener('resize', () => syncWideCardWidth(), { passive: true })
    observer.observe(document.documentElement, {
      subtree: true,
      childList: true,
      characterData: true,
      attributes: true,
      attributeFilter: ['placeholder', 'title', 'aria-label'],
    })
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', start, { once: true })
  } else {
    start()
  }
})()

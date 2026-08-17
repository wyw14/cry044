import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import type { Component } from 'vue'
import Standards from '../pages/Standards.vue'
import Compare from '../pages/Compare.vue'
import Workspace from '../pages/Workspace.vue'
import Disagreements from '../pages/Disagreements.vue'
import History from '../pages/History.vue'
import Approval from '../pages/Approval.vue'
import Transfer from '../pages/Transfer.vue'

type ReviewArea = 'standards' | 'deliberation' | 'audit' | 'approval' | 'exchange'

function screen(path: string, name: string, component: Component, area: ReviewArea): RouteRecordRaw {
  return { path, name, component, meta: { area, shell: 'review-standard-hall' } }
}

const reviewRoutes = [
  screen('/', 'standard-lineage', Standards, 'standards'),
  screen('/compare', 'template-diff', Compare, 'standards'),
  screen('/work', 'review-seat', Workspace, 'deliberation'),
  screen('/disagreements', 'deliberation', Disagreements, 'deliberation'),
  screen('/history', 'batch-history', History, 'audit'),
  screen('/approval', 'publication-gate', Approval, 'approval'),
  screen('/transfer', 'controlled-exchange', Transfer, 'exchange'),
]

const reviewRouter = createRouter({ history: createWebHistory(), routes: reviewRoutes })
reviewRouter.afterEach(() => window.scrollTo({ top: 0, behavior: 'instant' }))

export default reviewRouter

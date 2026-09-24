import { mount } from 'svelte'
import './app.css'
import './prose.css'
import DocsApp from './docs/DocsApp.svelte'

mount(DocsApp, { target: document.getElementById('app')! })

import { mount } from 'svelte'
import './app.css'
import Home from './Home.svelte'

mount(Home, { target: document.getElementById('app')! })

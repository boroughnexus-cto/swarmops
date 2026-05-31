export { matchers } from './matchers.js';

export const nodes = [
	() => import('./nodes/0'),
	() => import('./nodes/1'),
	() => import('./nodes/2'),
	() => import('./nodes/3'),
	() => import('./nodes/4'),
	() => import('./nodes/5'),
	() => import('./nodes/6'),
	() => import('./nodes/7'),
	() => import('./nodes/8'),
	() => import('./nodes/9'),
	() => import('./nodes/10'),
	() => import('./nodes/11'),
	() => import('./nodes/12'),
	() => import('./nodes/13'),
	() => import('./nodes/14'),
	() => import('./nodes/15'),
	() => import('./nodes/16'),
	() => import('./nodes/17'),
	() => import('./nodes/18'),
	() => import('./nodes/19'),
	() => import('./nodes/20')
];

export const server_loads = [];

export const dictionary = {
		"/": [2],
		"/agents": [3],
		"/dev-server/[directoryId]": [4],
		"/directories": [5],
		"/files": [6],
		"/files/[directoryId]": [7],
		"/git/[directoryId]": [8],
		"/kanban": [9],
		"/login": [10],
		"/projects": [11],
		"/projects/[id]": [12],
		"/settings": [13],
		"/swarm": [14],
		"/swarm/live": [16],
		"/swarm/[session]": [15],
		"/task-executions": [17],
		"/task-executions/[id]": [18],
		"/terminal": [19],
		"/terminal/[session]": [20]
	};

export const hooks = {
	handleError: (({ error }) => { console.error(error) }),
	
	reroute: (() => {}),
	transport: {}
};

export const decoders = Object.fromEntries(Object.entries(hooks.transport).map(([k, v]) => [k, v.decode]));

export const hash = false;

export const decode = (type, value) => decoders[type](value);

export { default as root } from '../root.js';
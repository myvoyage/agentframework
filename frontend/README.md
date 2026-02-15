# AgentFramework Desktop Frontend

This is the frontend for the AgentFramework desktop application, built with Vue 3 + TypeScript + Vite + Element Plus.

## 📋 Features

- **Workflow Editor**: Visual workflow editor with drag-and-drop support for creating and editing workflows
- **Skill Manager**: Manage built-in and custom skills
- **Config Manager**: Configure the AgentFramework settings
- **File Explorer**: Browse and manage files with security sandboxing
- **Responsive Design**: Adaptive layout for different screen sizes
- **Real-time Execution**: Visualize workflow execution with real-time status updates

## 🛠️ Tech Stack

- **Vue 3**: Progressive JavaScript framework
- **TypeScript**: Type-safe JavaScript
- **Vite**: Fast build tool and dev server
- **Element Plus**: UI component library
- **G6**: Graph visualization library for workflow editor
- **Pinia**: State management
- **Vue Router**: Routing

## 📁 Directory Structure

```
frontend/
├── src/
│   ├── assets/          # Static assets (images, fonts)
│   ├── router/         # Vue Router configuration
│   ├── views/           # Page components
│   │   ├── ConfigManager.vue      # Configuration management
│   │   ├── FileExplorer.vue       # File system browser
│   │   ├── Home.vue               # Home page
│   │   ├── SkillManager.vue       # Skill management
│   │   └── WorkflowEditor.vue     # Workflow editor
│   ├── App.vue          # Root component
│   ├── main.ts          # Application entry point
│   └── style.css        # Global styles
├── wailsjs/             # Wails generated bindings
├── index.html           # HTML template
├── package.json         # Dependencies and scripts
├── tsconfig.json        # TypeScript configuration
└── vite.config.ts       # Vite configuration
```

## 🚀 Development

### Prerequisites

- Node.js 16+
- NPM 8+

### Installation

```bash
cd frontend
npm install
```

### Build

```bash
npm run build
```

### Development

The frontend is typically developed alongside the backend using Wails dev mode:

```bash
# From project root
wails dev
```

## 🔗 Backend Integration

The frontend communicates with the backend through Wails-generated bindings in the `wailsjs` directory. These bindings
provide type-safe access to Go functions defined in `app.go`.

## 📚 Documentation

- [Vue 3 Documentation](https://vuejs.org/)
- [TypeScript Documentation](https://www.typescriptlang.org/)
- [Vite Documentation](https://vitejs.dev/)
- [Element Plus Documentation](https://element-plus.org/)
- [G6 Documentation](https://g6.antv.antgroup.com/)
- [Wails Documentation](https://wails.io/)


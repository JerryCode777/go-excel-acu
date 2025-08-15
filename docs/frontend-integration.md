# Frontend Integration Guide

## 🌐 Integración con React

Esta guía explica cómo integrar el frontend React con la API de GoExcel.

## 📦 Setup inicial

### Dependencias recomendadas
```bash
npm install axios
npm install @types/node  # si usas TypeScript
```

### Cliente API
```javascript
// api/client.js
import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export default apiClient;
```

## 🏗️ Componentes principales

### 1. Editor ACU con validación
```jsx
// components/ACUEditor.jsx
import React, { useState, useCallback } from 'react';
import { validateACU, createProject } from '../api/projects';

function ACUEditor() {
  const [acuContent, setAcuContent] = useState('');
  const [isValid, setIsValid] = useState(true);
  const [validationMessage, setValidationMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  // Validación en tiempo real (debounced)
  const validateContent = useCallback(
    debounce(async (content) => {
      if (!content.trim()) return;
      
      try {
        const result = await validateACU(content);
        setIsValid(result.valid);
        setValidationMessage(result.message);
      } catch (error) {
        setIsValid(false);
        setValidationMessage('Error de conexión');
      }
    }, 500),
    []
  );

  const handleContentChange = (e) => {
    const content = e.target.value;
    setAcuContent(content);
    validateContent(content);
  };

  const handleSubmit = async () => {
    if (!isValid) return;

    setIsLoading(true);
    try {
      // Convertir .acu a JSON en frontend
      const jsonData = parseACUToJSON(acuContent);
      
      // Enviar al backend
      const result = await createProject(jsonData);
      
      console.log('Proyecto creado:', result);
      // Redirect o mostrar success message
      
    } catch (error) {
      console.error('Error creando proyecto:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="acu-editor">
      <div className="editor-header">
        <h2>Editor ACU</h2>
        <div className={`validation-status ${isValid ? 'valid' : 'invalid'}`}>
          {validationMessage}
        </div>
      </div>
      
      <textarea
        value={acuContent}
        onChange={handleContentChange}
        placeholder="Pega tu código .acu aquí o escríbelo manualmente..."
        rows={25}
        cols={100}
        className={`acu-textarea ${!isValid ? 'error' : ''}`}
      />
      
      <div className="editor-actions">
        <button 
          onClick={handleSubmit}
          disabled={!isValid || isLoading}
          className="btn-primary"
        >
          {isLoading ? 'Creando...' : 'Crear Proyecto'}
        </button>
      </div>
    </div>
  );
}

// Utility function para debounce
function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}
```

### 2. Parser ACU en JavaScript
```javascript
// utils/acuParser.js

export function parseACUToJSON(acuContent) {
  // Parser básico para convertir .acu a JSON
  const result = {
    proyecto: {},
    partidas: []
  };

  // Extraer proyecto
  const projectMatch = acuContent.match(/@proyecto\s*\{([^,]+),\s*([^}]+)\}/);
  if (projectMatch) {
    const projectFields = parseFields(projectMatch[2]);
    result.proyecto = {
      nombre: cleanQuotes(projectFields.nombre || ''),
      descripcion: cleanQuotes(projectFields.descripcion || ''),
      moneda: cleanQuotes(projectFields.moneda || 'PEN')
    };
  }

  // Extraer partidas
  const partidaMatches = acuContent.matchAll(/@partida\s*\{([^,]+),\s*((?:[^{}]*\{[^{}]*\}[^{}]*)*[^}]*)\}/g);
  
  for (const match of partidaMatches) {
    const partidaId = match[1].trim();
    const partidaContent = match[2];
    
    const partida = parsePartida(partidaContent);
    if (partida.codigo) {
      result.partidas.push(partida);
    }
  }

  return result;
}

function parsePartida(content) {
  const fields = parseFields(content);
  
  const partida = {
    codigo: cleanQuotes(fields.codigo || ''),
    descripcion: cleanQuotes(fields.descripcion || ''),
    unidad: cleanQuotes(fields.unidad || ''),
    rendimiento: parseFloat(fields.rendimiento || 0),
    mano_obra: parseRecursos(content, 'mano_obra'),
    materiales: parseRecursos(content, 'materiales'),
    equipos: parseRecursos(content, 'equipos'),
    subcontratos: parseRecursos(content, 'subcontratos')
  };

  return partida;
}

function parseRecursos(content, tipoRecurso) {
  const pattern = new RegExp(`${tipoRecurso}\\s*=\\s*\\{([^}]+)\\}`);
  const match = content.match(pattern);
  
  if (!match) return [];

  const recursosContent = match[1];
  const recursoMatches = recursosContent.matchAll(/\{([^}]+)\}/g);
  
  const recursos = [];
  for (const recursoMatch of recursoMatches) {
    const recursoFields = parseFields(recursoMatch[1]);
    
    const recurso = {
      codigo: cleanQuotes(recursoFields.codigo || ''),
      descripcion: cleanQuotes(recursoFields.desc || ''),
      unidad: cleanQuotes(recursoFields.unidad || ''),
      cantidad: parseFloat(recursoFields.cantidad || 0),
      precio: parseFloat(recursoFields.precio || 0)
    };

    if (recursoFields.cuadrilla) {
      recurso.cuadrilla = parseFloat(recursoFields.cuadrilla);
    }

    if (recurso.codigo) {
      recursos.push(recurso);
    }
  }

  return recursos;
}

function parseFields(content) {
  const fields = {};
  const fieldRegex = /(\w+)\s*=\s*("([^"]*)"|([^,}{\s]+))/g;
  let match;

  while ((match = fieldRegex.exec(content)) !== null) {
    const key = match[1];
    const value = match[3] || match[4];
    fields[key] = value;
  }

  return fields;
}

function cleanQuotes(value) {
  return value.replace(/^["']|["']$/g, '');
}
```

### 3. API Service
```javascript
// api/projects.js
import apiClient from './client';

export const getProjects = async () => {
  const response = await apiClient.get('/projects');
  return response.data;
};

export const getProject = async (id) => {
  const response = await apiClient.get(`/projects/${id}`);
  return response.data;
};

export const createProject = async (projectData) => {
  const response = await apiClient.post('/projects', projectData);
  return response.data;
};

export const updateProject = async (id, projectData) => {
  const response = await apiClient.put(`/projects/${id}`, projectData);
  return response.data;
};

export const deleteProject = async (id) => {
  const response = await apiClient.delete(`/projects/${id}`);
  return response.data;
};

export const exportProject = async (id, format = 'excel') => {
  const response = await apiClient.get(`/projects/${id}/export`, {
    params: { format },
    responseType: format === 'excel' ? 'blob' : 'json'
  });
  return response.data;
};

export const validateACU = async (acuContent) => {
  const response = await apiClient.post('/validate-acu', {
    acu_content: acuContent
  });
  return response.data;
};
```

### 4. Lista de proyectos
```jsx
// components/ProjectList.jsx
import React, { useState, useEffect } from 'react';
import { getProjects, deleteProject } from '../api/projects';

function ProjectList() {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadProjects();
  }, []);

  const loadProjects = async () => {
    try {
      const data = await getProjects();
      setProjects(data.projects);
    } catch (error) {
      console.error('Error loading projects:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('¿Estás seguro de eliminar este proyecto?')) return;
    
    try {
      await deleteProject(id);
      loadProjects(); // Reload list
    } catch (error) {
      console.error('Error deleting project:', error);
    }
  };

  if (loading) return <div>Cargando proyectos...</div>;

  return (
    <div className="project-list">
      <h2>Mis Proyectos</h2>
      {projects.length === 0 ? (
        <p>No hay proyectos creados.</p>
      ) : (
        <div className="projects-grid">
          {projects.map(project => (
            <div key={project.id} className="project-card">
              <h3>{project.nombre}</h3>
              <p>{project.descripcion}</p>
              <div className="project-meta">
                <span>Moneda: {project.moneda}</span>
                <span>Creado: {project.created_at}</span>
              </div>
              <div className="project-actions">
                <button onClick={() => window.open(`/projects/${project.id}`)}>
                  Ver
                </button>
                <button onClick={() => handleDelete(project.id)}>
                  Eliminar
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
```

## 🎨 Estilos recomendados

### CSS para editor ACU
```css
/* styles/ACUEditor.css */
.acu-editor {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.validation-status {
  padding: 8px 16px;
  border-radius: 4px;
  font-weight: bold;
}

.validation-status.valid {
  background-color: #d4edda;
  color: #155724;
}

.validation-status.invalid {
  background-color: #f8d7da;
  color: #721c24;
}

.acu-textarea {
  width: 100%;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  border: 2px solid #ddd;
  border-radius: 8px;
  padding: 16px;
  resize: vertical;
}

.acu-textarea.error {
  border-color: #dc3545;
}

.acu-textarea:focus {
  outline: none;
  border-color: #007bff;
}

.editor-actions {
  margin-top: 20px;
  text-align: right;
}

.btn-primary {
  background-color: #007bff;
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
}

.btn-primary:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}
```

## 🔗 Rutas recomendadas

```jsx
// App.js
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/projects" element={<ProjectList />} />
        <Route path="/projects/new" element={<ACUEditor />} />
        <Route path="/projects/:id" element={<ProjectDetail />} />
        <Route path="/projects/:id/edit" element={<ACUEditor />} />
      </Routes>
    </Router>
  );
}
```

## ⚡ Consejos de performance

1. **Debounce validation**: Evita validar en cada keystroke
2. **Parse caching**: Cachea resultados de parsing pesados
3. **Lazy loading**: Carga componentes bajo demanda
4. **Error boundaries**: Maneja errores de parsing gracefully

## 🔍 Testing

```javascript
// __tests__/acuParser.test.js
import { parseACUToJSON } from '../utils/acuParser';

test('parses basic ACU content', () => {
  const acu = `
    @proyecto{test,
      nombre = "Test Project",
      moneda = "PEN"
    }
    
    @partida{excavacion,
      codigo = "01.01.01",
      descripcion = "EXCAVACIÓN MANUAL",
      unidad = "m3",
      rendimiento = 8.0
    }
  `;
  
  const result = parseACUToJSON(acu);
  
  expect(result.proyecto.nombre).toBe('Test Project');
  expect(result.partidas).toHaveLength(1);
  expect(result.partidas[0].codigo).toBe('01.01.01');
});
```
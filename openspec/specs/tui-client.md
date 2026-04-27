# tui-client Specification

## Descripción y responsabilidad
El `tui-client` es la Interfaz de Usuario de Terminal (TUI) desarrollada en Go utilizando Bubbletea. Actúa como el cliente principal para usuarios técnicos, ofreciendo una experiencia tipo chat inmersiva donde el usuario interactúa directamente con el `ai-agent`.

## Requisitos funcionales
1. El cliente MUST renderizar una interfaz de chat a pantalla completa al ejecutar `neto`.
2. El cliente MUST permitir entrada de texto libre por parte del usuario y mostrar el historial de la conversación.
3. El cliente MUST gestionar el flujo de autenticación al primer uso, proveyendo un mecanismo de login y almacenando un token JWT local.
4. El cliente MUST enviar y recibir mensajes hacia el Agente IA en tiempo real o simulando carga con loading spinners.
5. El cliente MUST renderizar feedback visual enriquecido de las acciones (ej. tablas resumen de saldo, notificaciones de éxito/error).
6. El cliente MUST soportar navegación básica con teclado (ej. vistas de chat, ayuda, configuración).

## Requisitos no funcionales
- **Performance:** MUST tener un arranque rápido.
- **Usabilidad:** MUST utilizar convenciones de terminal (colores, atajos de teclado para la navegación).
- **Cross-platform:** MUST ejecutarse correctamente en Linux, macOS y Windows.

## Entidades/contratos de datos relevantes
- `LocalSession`: Token de Auth, IDUsuario, Preferencias de vista guardadas localmente.
- `ViewModel`: Estado de Bubbletea (mensajes, input actual, viewport, estado de carga).

## Escenarios clave

#### Scenario: Primer uso sin autenticación
- GIVEN que el usuario abre `neto` por primera vez
- WHEN el cliente no encuentra un token local
- THEN debe mostrar una pantalla de Login solicitando credenciales
- AND una vez validado, debe guardar el token y pasar a la vista principal de chat.

#### Scenario: Interacción visual tras comando
- GIVEN el usuario está en la vista principal y escribe "Mostrar saldo total"
- WHEN el agente retorna una respuesta exitosa con los datos financieros
- THEN el cliente debe renderizar el texto del agente 
- AND opcionalmente un componente visual de Bubbletea (ej. tabla formateada) mostrando el saldo por cuenta.

## Fuera de scope explícito
- Renderizado de gráficos complejos (charts no ASCII avanzados).
- Panel de administración web.

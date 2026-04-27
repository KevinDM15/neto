# ai-agent Specification

## Descripción y responsabilidad
El `ai-agent` es el intérprete de lenguaje natural motorizado por Anthropic Claude. Se encarga de traducir la intención del usuario a comandos de negocio accionables, ejecutando herramientas (tool use) para mutar el estado o consultar datos a través del `core-domain` y la infraestructura (Supabase).

## Requisitos funcionales
1. El sistema MUST interpretar lenguaje natural para extraer intenciones financieras (registrar gastos, consultar saldo, etc.).
2. El sistema MUST mapear las intenciones a llamadas a funciones (tools) predefinidas (ej. `crear_transaccion`, `listar_gastos`).
3. El sistema MUST pedir confirmación explícita al usuario antes de ejecutar tools que realicen mutaciones destructivas o alteraciones importantes (ej. borrar transacciones).
4. El sistema MUST ejecutar tools con el contexto de seguridad correcto (aislamiento multi-tenant por usuario).
5. El sistema MUST manejar errores de validación de negocio informando al usuario en lenguaje natural.

## Requisitos no funcionales
- **Seguridad (RLS):** MUST asegurar que toda acción generada por el LLM se ejecuta bajo la identidad del usuario autenticado.
- **Tolerancia a fallos:** SHOULD manejar tiempos de respuesta largos de la API de Anthropic o caídas del servicio de manera resiliente.
- **Claridad:** La respuesta del agente MUST ser concisa y confirmatoria.

## Entidades/contratos de datos relevantes
- `Mensaje`: ID, IDUsuario, Role (user/assistant/system), Contenido, Fecha.
- `ToolCall`: Nombre, Parametros (JSON estructurado según las entidades del core).
- `ContextoAgente`: IDUsuario, FechaActual, MonedaPrincipal, Lista de Cuentas (para enriquecer el prompt).

## Escenarios clave

#### Scenario: Registro de gasto exitoso
- GIVEN el usuario dice "Gasté 50 en la cena con tarjeta de crédito"
- WHEN el agente interpreta el comando
- THEN el agente debe invocar la tool `crear_transaccion` con monto=50, cuenta="tarjeta de crédito", categoria="Comida/Cena"
- AND debe responder "He registrado un gasto de $50 en Cena pagado con tarjeta de crédito."

#### Scenario: Mutación riesgosa requiere confirmación
- GIVEN el usuario dice "Borra todos mis gastos de ayer"
- WHEN el agente analiza la intención destructiva
- THEN el agente NO DEBE ejecutar la acción inmediatamente
- AND debe responder pidiendo confirmación: "¿Estás seguro de que quieres eliminar [N] gastos registrados ayer?"

## Fuera de scope explícito
- Entrenamiento fino de un modelo propio (se usa Claude pre-entrenado).
- Interfaces visuales para mostrar el chat (eso va en TUI/Web).

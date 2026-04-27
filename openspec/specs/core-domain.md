# core-domain Specification

## Descripción y responsabilidad
El `core-domain` contiene las entidades puras y las reglas de negocio agnósticas a cualquier framework, interfaz o base de datos. Es el núcleo de Neto y dicta cómo se manejan y validan las finanzas personales (usuarios, cuentas, transacciones, presupuestos, categorías, monedas).

## Requisitos funcionales
1. El sistema MUST gestionar cuentas con un saldo calculado derivado de sus transacciones.
2. El sistema MUST estructurar categorías como un árbol jerárquico (padre/hijo).
3. El sistema MUST permitir la creación de presupuestos, asociándolos a categorías y rangos de fechas.
4. El sistema MUST calcular la diferencia entre gasto real y presupuesto asignado.
5. El sistema MUST manejar listas de monedas predefinidas (LatAm + USD + EUR) como seeds, permitiendo conversiones de divisas.
6. El sistema MUST manejar conceptos de "Metas" de ahorro y "Deudas", con su respectivo progreso de liquidación.

## Requisitos no funcionales
- **Agnóstico:** NO MUST depender de frameworks externos, HTTP, o bases de datos (arquitectura limpia).
- **Testabilidad:** MUST poseer cobertura de test unitaria alta en sus reglas lógicas.
- **Precisión:** MUST usar tipos seguros para dinero (ej. enteros/cents) evitando errores de coma flotante.

## Entidades/contratos de datos relevantes
- `Usuario`: ID, Nombre, Email, Preferencias.
- `Cuenta`: ID, IDUsuario, Nombre, Tipo (banco, efectivo, crédito), IDMoneda.
- `Transaccion`: ID, IDCuenta, IDCategoria, Monto, Fecha, Tipo (ingreso, egreso, transferencia).
- `Categoria`: ID, IDUsuario, Nombre, IDPadre (para jerarquía).
- `Presupuesto`: ID, IDUsuario, IDCategoria, Limite, Periodo.
- `Moneda`: Codigo (ISO), Simbolo, TasaCambio.

## Escenarios clave

#### Scenario: Categorización y balance de transacciones
- GIVEN que una cuenta en USD tiene un saldo de 1000
- WHEN se crea una nueva transacción de egreso por 200 asignada a la categoría "Comida"
- THEN el balance de la cuenta debe ser 800
- AND el gasto total en "Comida" del mes debe aumentar en 200

#### Scenario: Control de Presupuesto
- GIVEN un presupuesto mensual de 500 para "Transporte"
- WHEN el gasto acumulado mensual en "Transporte" llega a 550
- THEN el sistema debe calcular un desvío (overspend) de -50

## Fuera de scope explícito
- Integración real con APIs de tipos de cambio (pertenece a infraestructura).
- Lógica de la base de datos o almacenamiento en Supabase.
- Interfaces de usuario o endpoints.

📚 *EJEMPLOS DE COMANDOS - BOT DE GASTOS EN PAREJA*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💰 *AGREGAR GASTOS* (`/add`)

Formato: `/add <monto> <descripción> [categoría] [método_pago]`

Ejemplos básicos:
• `/add 50.00 Supermercado`
• `/add 1250.50 Alquiler`
• `/add 25.50 Cena Restaurante`

Con categoría:
• `/add 50.00 Supermercado Comida`
• `/add 500 Netflix Servicios`

Con método de pago:
• `/add 50.00 Supermercado Comida Visa`
• `/add 25.50 Cena Efectivo`

Para tu pareja:
• `/add 50.00 Supermercado Comida Visa pareja`
• `/add 25.50 Cena partner`

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✏️ *EDITAR GASTOS* (`/edit`)

Formato: `/edit <id_gasto> <campo> <valor>`

Campos: `category`, `payment_method` (o `payment`)

Ejemplos:
• `/edit 123 category Supermercado`
• `/edit 456 payment_method Visa`
• `/edit 789 payment Efectivo`

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🗑️ *ELIMINAR GASTOS* (`/delete`)

Formato: `/delete <id_gasto>`

Ejemplos:
• `/delete 123`
• `/delete 456`

Sin ID muestra últimos 10 gastos:
• `/delete`

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💳 *AGREGAR MÉTODOS DE PAGO* (`/payment_methods add`)

Formato: `/payment_methods add <nombre> <tipo> [día_cierre]`

Tipos aceptados (inglés o español):
• `credit_card` / `TarjetaCredito` / `tarjeta_credito`
• `debit_card` / `TarjetaDebito` / `tarjeta_debito`
• `cash` / `Efectivo`
• `bank_transfer` / `Transferencia`
• `other` / `Otro`

Ejemplos:
• `/payment_methods add Visa credit_card 15`
• `/payment_methods add Visa TarjetaCredito 15`
• `/payment_methods add Debito TarjetaDebito`
• `/payment_methods add Efectivo cash`
• `/payment_methods add Transferencia bank_transfer`
• `/payment_methods add PayPal other`

⚠️ Las tarjetas de crédito requieren día de cierre (1-31)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *VER MÉTODOS DE PAGO*
• `/payment_methods`

🔧 *EDITAR MÉTODOS DE PAGO*
• `/payment_methods edit <id> <campo> <valor>`
• Ejemplo: `/payment_methods edit 1 closing_day 20`

🗑️ *ELIMINAR MÉTODOS DE PAGO*
• `/payment_methods delete <id>`
• Ejemplo: `/payment_methods delete 1`

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *NOTAS:*
• Montos con punto: `50.00`, `1250.50`
• Los IDs se muestran al crear/listar
• Nombres de métodos de pago no distinguen mayúsculas
• Usá `/help` para ver todos los comandos

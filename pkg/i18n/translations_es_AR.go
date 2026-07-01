package i18n

var spanishARTranslations = map[string]string{
	// Welcome messages
	"welcome":             "👋 ¡Bienvenido al Bot de Gastos en Pareja!\n\n¡Hola %s! Te voy a ayudar a vos y a tu pareja a llevar los gastos compartidos.\n\nPara empezar:\n1. Creá un lobby con /start\n2. Compartí el código del lobby con tu pareja\n3. Empezá a agregar gastos con /add\n\nUsá /help para ver todos los comandos disponibles.",
	"welcome_back":        "👋 ¡Bienvenido de nuevo, %s!\n\n*Tu Lobby:*\nID del Lobby: `%d`\nTipo de Cuenta: %s\n%s\n\nUsá /help para ver todos los comandos disponibles.",
	"lobby_created":       "👋 ¡Bienvenido al Bot de Gastos en Pareja!\n\n¡Hola %s! Creé un nuevo lobby para vos.\n\n*Detalles de tu Lobby:*\nID del Lobby: `%d`\nTipo de Cuenta: %s\n\n*Próximos Pasos:*\n1. Compartí este token de invitación con tu pareja: `%s`\n2. Tu pareja debería ejecutar: `/start %s`\n3. Una vez que ambos estén, empezá a agregar gastos con `/add`\n\nUsá /help para ver todos los comandos disponibles.",
	"lobby_created_group": "👋 ¡Bienvenido al Bot de Gastos en Pareja!\n\n¡Hola %s! Creé un lobby para este grupo.\n\n*Detalles del Lobby:*\nID del Lobby: `%d`\nTipo de Cuenta: %s\n\nTu pareja puede unirse ejecutando `/start` en este grupo.\n\nUsá /help para ver todos los comandos disponibles.",
	"lobby_ready_group":   "✅ ¡El lobby está listo!\n\n*Detalles del Lobby:*\nID del Lobby: `%d`\nTipo de Cuenta: %s\n%s\n\nYa podés empezar a agregar gastos con `/add`\n\nUsá /help para ver todos los comandos disponibles.",
	"lobby_joined":        "✅ ¡Te uniste exitosamente al lobby %d!",
	"lobby_joined_token":  "✅ ¡Te uniste exitosamente al lobby!",
	"waiting_partner":     "Esperando que se una tu pareja...",
	"partner_id":          "ID de Pareja: %d",
	"lobby_security_info": "🔒 *Información de Seguridad:*\n\nTu lobby está protegido por un token de invitación. Compartí este token SOLO con tu pareja:\n\n`%s`\n\n*Cómo unirse:*\nTu pareja debería ejecutar:\n`/start %s`\n\n⚠️ ¡Mantené este token privado! Cualquiera con este token puede unirse a tu lobby.",

	// Error messages
	"error_user_init":        "❌ Error: No se pudo inicializar el usuario. Por favor intentá de nuevo.",
	"error_lobby_check":      "❌ Error: No se pudo verificar el estado del lobby. Por favor intentá de nuevo.",
	"error_lobby_not_found":  "❌ Todavía no estás en un lobby. Usá /start para crear o unirte a uno.",
	"error_lobby_join":       "❌ No se pudo unir al lobby: %v",
	"error_lobby_create":     "❌ Error: No se pudo crear el lobby. Por favor intentá de nuevo.",
	"error_invalid_lobby_id": "❌ Token de invitación inválido. Uso: `/start <invite_token>` para unirte a un lobby existente.",
	"error_invalid_token":    "❌ Token de invitación inválido o expirado. Por favor pedile a tu pareja un nuevo token.",
	"error_unknown_command":  "Comando desconocido. Usá /help para ver los comandos disponibles.",
	"error_invalid_user_id":  "❌ ID de usuario inválido. Usá 'user1', 'user2', 'partner', o un ID de usuario válido de tu lobby.",
	"error_generic":          "❌ Error: %v",
	"error_invalid_period":   "❌ Formato de período inválido. Usá YYYY-MM",

	// Help
	"help": `📚 *Comandos Disponibles:*

*Comandos Básicos:*
/start - Inicializar bot y crear/unirse a lobby
/help - Mostrar este mensaje de ayuda
/examples - Mostrar ejemplos de uso de comandos

*Gestión de Gastos:*
/add <monto> <descripción> [categoría] [método_pago] - Agregar un gasto
/list [mes] - Listar gastos (mes actual o especificado)
/list_billing [método_pago] [período] - Listar gastos por ciclo de facturación
/delete [id_gasto] - Eliminar un gasto (muestra gastos recientes si no se proporciona ID)
/edit <id_gasto> <campo> <valor> - Editar un gasto (campos: category, payment_method)

*Reportes y Análisis:*
/summary [fecha_inicio] [fecha_fin] - Obtener resumen de gastos
/summary_billing [método_pago] [período] - Obtener resumen por ciclo de facturación
/settle - Calcular quién le debe a quién
/settle_billing [método_pago] [período] - Calcular liquidación para período de facturación
/export_ynab [mes] - Exportar gastos como archivo TSV compatible con YNAB (para importar en Gasti)

*Configuración:*
/payment_methods - Gestionar métodos de pago (agregar, editar, eliminar)
  Ejemplos:
  ` + "`/payment_methods`" + ` - Listar todos los métodos de pago
  ` + "`/payment_methods add Visa credit_card 15`" + ` - Agregar tarjeta de crédito con día de cierre 15
  ` + "`/payment_methods edit 1 closing_day 20`" + ` - Editar método de pago #1
  ` + "`/payment_methods delete 1`" + ` - Eliminar método de pago #1

/categories - Gestionar categorías
  Ejemplo: ` + "`/categories`" + ` - Listar todas las categorías

/settings - Configurar tipo de cuenta, porcentajes de sueldo
  Ejemplos:
  ` + "`/settings`" + ` - Mostrar configuración actual
  ` + "`/settings account_type shared`" + ` - Establecer tipo de cuenta compartida
  ` + "`/settings salary 0.6 0.4`" + ` - Establecer porcentajes de sueldo (60% usuario1, 40% usuario2)

/language - Cambiar idioma
  Ejemplos:
  ` + "`/language`" + ` - Mostrar idioma actual
  ` + "`/language en`" + ` - Cambiar a Inglés
  ` + "`/language es_AR`" + ` - Cambiar a Español

*Ejemplos:*
` + "`/add 50.00 Supermercado`" + `
` + "`/add 25.50 Cena tarjeta_1`" + `
` + "`/summary 2024-01-01 2024-01-31`" + `
` + "`/settle`" + `

Para más detalles, usá cada comando sin argumentos para ver su uso.`,

	// Settings
	"settings_current":      "⚙️ *Configuración Actual del Lobby*\n\nID del Lobby: `%d`\nTipo de Cuenta: `%s`\nSueldo Usuario 1 %%: %.1f%%\nSueldo Usuario 2 %%: %.1f%%\n\n*Para cambiar la configuración:*\n`/settings account_type <separate|shared>`\n`/settings salary <user1_pct> <user2_pct>`\n\nEjemplo:\n`/settings account_type shared`\n`/settings salary 0.6 0.4`",
	"settings_updated":      "✅ ¡Configuración actualizada exitosamente!",
	"settings_usage":        "❌ Uso: `/settings account_type <separate|shared>`",
	"settings_invalid_type": "❌ El tipo de cuenta debe ser 'separate' o 'shared'",
	"settings_salary_usage": "❌ Uso: `/settings salary <porcentaje_user1> <porcentaje_user2>`\nEjemplo: `/settings salary 0.6 0.4`",
	"settings_invalid_pct":  "❌ Valores de porcentaje inválidos. Usá números entre 0 y 1.",
	"settings_pct_range":    "❌ Los porcentajes deben estar entre 0 y 1.",
	"settings_unknown":      "❌ Configuración desconocida. Usá `account_type` o `salary`.",
	"settings_error":        "❌ No se pudo actualizar la configuración: %v",

	// Payment methods
	"payment_methods_none":            "📋 No hay métodos de pago configurados.\n\nAgregá uno con:\n`/payment_methods add <nombre> <tipo> [día_cierre]`\n\nTipos: credit_card (o TarjetaCredito), debit_card (o TarjetaDebito), cash (o Efectivo), bank_transfer (o Transferencia), other (o Otro)",
	"payment_methods_list":            "📋 *Métodos de Pago:*\n\n%s",
	"payment_method_item":             "%s *%s* (%s)",
	"payment_method_closing":          " - Cierra el día %d",
	"payment_method_owner":            " - Dueño: %d",
	"payment_method_added":            "✅ ¡Método de pago *%s* creado exitosamente!",
	"payment_method_closing_day":      "\nDía de cierre: %d",
	"payment_method_add_usage":        "❌ Uso: `/payment_methods add <nombre> <tipo> [día_cierre]`\n\nTipos: credit_card (o TarjetaCredito), debit_card (o TarjetaDebito), cash (o Efectivo), bank_transfer (o Transferencia), other (o Otro)\nEjemplo: `/payment_methods add Visa credit_card 15` o `/payment_methods add Visa TarjetaCredito 15`",
	"payment_method_closing_required": "❌ Las tarjetas de crédito requieren un día de cierre. Uso: `/payment_methods add <nombre> credit_card <día_cierre>`",
	"payment_method_closing_invalid":  "❌ El día de cierre debe ser un número entre 1 y 31",
	"payment_method_not_found":        "⚠️ Método de pago '%s' no encontrado.",
	"payment_method_not_found_list":   "⚠️ Método de pago '%s' no encontrado.\n\nMétodos disponibles:\n%s\n\nGasto agregado sin método de pago.",
	"payment_method_add_error":        "❌ No se pudo crear el método de pago: %v",
	"payment_method_edit_usage":       "❌ Uso: `/payment_methods edit <id> <campo> <valor>`\n\nCampos: name, type, closing_day, active\nEjemplo: `/payment_methods edit 1 closing_day 20`",
	"payment_method_delete_usage":     "❌ Uso: `/payment_methods delete <id>`",
	"payment_method_invalid_id":       "❌ ID de método de pago inválido",
	"payment_method_update_error":     "❌ No se pudo actualizar el método de pago: %v",
	"payment_method_delete_error":     "❌ No se pudo eliminar el método de pago: %v",
	"payment_method_updated":          "✅ ¡Método de pago actualizado exitosamente!",
	"payment_method_deleted":          "✅ ¡Método de pago eliminado exitosamente!",
	"payment_method_unknown_action":   "❌ Acción desconocida. Usá: `add`, `edit`, o `delete`",

	// Expenses
	"expense_add_usage":           "❌ Uso: `/add <monto> <descripción> [categoría] [método_pago] [fecha]`\n\nEjemplos:\n`/add 50.00 Supermercado`\n`/add 25.50 Cena tarjeta_1`\n`/add 30.00 Almuerzo 2024-01-15`\n\nLa fecha puede ser en formato: 2024-01-15, 15/01/2024, etc.",
	"expense_add_cuotas_usage":    "❌ Uso: `/add_cuotas <monto_total> <descripción> <cuotas> [categoría] [método_pago] [quien_gastó] [fecha]`\n\nEjemplos:\n`/add_cuotas 120000 Celular 12 visa 2025-01-10`\n`/add_cuotas 30000 Zapas 3 ropa mastercard user2 2024-12-15`\n\nCrea N gastos mensuales (cuotas) vinculados entre sí.",
	"expense_invalid_installments": "❌ Cuotas inválidas. Por favor ingresá un número >= 2.",
	"expense_cuotas_added":        "✅ ¡Cuotas creadas!\n\nTotal: %s\nDescripción: %s\nCuotas: %d\nFecha de compra: %s\nID 1ra cuota: %d\n",
	"expense_invalid_amount":      "❌ Monto inválido. Por favor proporcioná un número positivo.",
	"expense_added":               "✅ ¡Gasto agregado!\n\nMonto: %s\nDescripción: %s\n",
	"expense_category":            "Categoría: %s\n",
	"expense_payment_method":      "Método de Pago: %s\n",
	"expense_billing_period":      "Período de Facturación: %s a %s\n",
	"expense_add_error":           "❌ No se pudo agregar el gasto: %v",
	"installments_created":        "✅ ¡Cuotas creadas!\n\nMonto total: %s\nDescripción: %s\nCuotas: %d\n\n",
	"expense_list_none":           "📋 No se encontraron gastos para %s.",
	"expense_list_header":         "📋 *Gastos* (%d)\n\n",
	"expense_list_item":           "• %s - %s\n",
	"expense_list_category":       "  Categoría: %s\n",
	"expense_list_date":           "  Fecha: %s\n\n",
	"expense_list_total":          "*Total: %s*",
	"expense_no_description":      "Sin descripción",
	"expense_billing_usage":       "❌ Uso: `/list_billing <método_pago> [período]`\n\nEjemplo: `/list_billing Visa 2024-01`",
	"expense_billing_no_cycle":    "❌ Este método de pago no tiene un ciclo de facturación configurado.",
	"expense_billing_none":        "📋 No se encontraron gastos para el período de facturación %s a %s.",
	"expense_billing_header":      "📋 *Gastos del Período de Facturación*\nMétodo de Pago: %s\nPeríodo: %s a %s\n\n",
	"expense_delete_usage":        "❌ Uso: `/delete <id_gasto>`\n\nEjemplo: `/delete 123`",
	"expense_delete_none":         "📋 No se encontraron gastos para eliminar.",
	"expense_delete_list_header":  "📋 *Gastos Recientes (Últimos 10):*\n\n",
	"expense_delete_invalid_id":   "❌ ID de gasto inválido. Uso: `/delete <id_gasto>`",
	"expense_delete_not_found":    "❌ Gasto no encontrado o no pertenece a tu lobby.",
	"expense_delete_error":        "❌ No se pudo eliminar el gasto: %v",
	"expense_deleted":             "✅ ¡Gasto eliminado exitosamente!",
	"expense_edit_usage":          "❌ Uso: `/edit <id_gasto> <campo> <valor>`\n\nCampos: `category` (categoría), `payment_method` (método_pago)\n\nEjemplos:\n`/edit 123 category Supermercado`\n`/edit 123 payment_method Visa`",
	"expense_edit_invalid_id":     "❌ ID de gasto inválido. Uso: `/edit <id_gasto> <campo> <valor>`",
	"expense_edit_not_found":      "❌ Gasto no encontrado o no pertenece a tu lobby.",
	"expense_edit_category_usage": "❌ Uso: `/edit <id_gasto> category <nombre_categoría>`",
	"expense_edit_payment_usage":  "❌ Uso: `/edit <id_gasto> payment_method <nombre_método_pago>`",
	"expense_edit_invalid_field":  "❌ Campo inválido. Usá `category` o `payment_method`.",
	"expense_edit_error":          "❌ No se pudo editar el gasto: %v",
	"expense_edited":              "✅ ¡Gasto actualizado exitosamente!",

	// Settlement
	"settle_usage":          "❌ Uso: `/settle_billing <método_pago> [período]`\n\nEjemplo: `/settle_billing Visa 2024-01`",
	"settle_error":          "❌ Error al calcular la liquidación: %v",
	"settle_report":         "💰 *Reporte de Liquidación*\n\nPeríodo: %s\nTipo de Cuenta: %s\nTotal de Gastos: %s\n\n",
	"settle_separate":       "*Cuentas Separadas (División Igual):*\n\n",
	"settle_shared":         "*Cuenta Compartida (Basada en Sueldo):*\n\n",
	"settle_user1_spent":    "Usuario 1 Gastó: %s\n",
	"settle_user2_spent":    "Usuario 2 Gastó: %s\n\n",
	"settle_user1_expected": "Usuario 1 Esperado (%.1f%%): %s\n",
	"settle_user2_expected": "Usuario 2 Esperado (%.1f%%): %s\n\n",
	"settle_expected_per":   "Esperado por persona: %s\n\n",
	"settle_user1_owes":     "➡️ Usuario 1 le debe a Usuario 2: %s\n",
	"settle_user2_owes":     "➡️ Usuario 2 le debe a Usuario 1: %s\n",
	"settle_all_settled":    "✅ ¡Todo saldado! Sin deudas.\n",

	// Summary
	"summary_none":          "📊 *Resumen*\n\nNo se encontraron gastos para %s.",
	"summary_header":        "📊 *Resumen de Gastos*\n\nPeríodo: %s\nTotal de Gastos: %s\nCantidad de Gastos: %d\n\n",
	"summary_by_person":     "*Por Persona:*\n",
	"summary_user1":         "Usuario 1: %s (%.1f%%)\n",
	"summary_user2":         "Usuario 2: %s (%.1f%%)\n\n",
	"summary_by_category":   "*Por Categoría:*\n",
	"summary_category_item": "• %s: %s (%.1f%%)\n",
	"summary_by_payment":    "*Por Método de Pago:*\n",
	"summary_billing_usage": "❌ Uso: `/summary_billing <método_pago> [período]`\n\nEjemplo: `/summary_billing Visa 2024-01`",
	"summary_period":        "el período seleccionado",

	// Analysis
	"analyze_error":          "❌ Error al analizar los gastos: %v",
	"analyze_header":         "📈 *Análisis de Gastos Mensuales*\n\nPeríodo Actual: %s\nPeríodo Anterior: %s\n\n",
	"analyze_overall":        "*Gastos Generales:*\n",
	"analyze_current":        "Actual: %s\n",
	"analyze_previous":       "Anterior: %s\n",
	"analyze_increase":       "📈 Aumento: %.1f%%\n\n",
	"analyze_decrease":       "📉 Disminución: %.1f%%\n\n",
	"analyze_no_change":      "➡️ Sin cambios\n\n",
	"analyze_spikes":         "*⚠️ Picos de Gasto (>20%% de aumento):*\n",
	"analyze_spike_item":     "• %s: %s (+%.1f%%)\n",
	"analyze_new_categories": "*🆕 Categorías Nuevas:*\n",
	"analyze_new_category":   "• %s\n",
	"analyze_discontinued":   "*❌ Categorías Discontinuadas:*\n",
	"analyze_top_changes":    "*Principales Cambios por Categoría:*\n",
	"analyze_change_item":    "• %s: %s → %s (%.1f%%)\n",

	// Language
	"language_current": "🌐 Idioma actual: %s\n\nIdiomas disponibles:\n%s",
	"language_changed": "✅ Idioma cambiado a %s",
	"language_usage":   "❌ Uso: `/language <código>`\n\nIdiomas disponibles:\n%s",
	"language_invalid": "❌ Código de idioma inválido. Disponibles: %s",

	// Receipt OCR
	"receipt_llm_disabled": "❌ El OCR de recibos no está disponible en este momento.",
	"receipt_processing":   "⏳ Analizando recibo...",
	"receipt_failed":       "❌ No pude leer el recibo. Asegurate de que la foto sea clara y tenga el monto visible.\n\nPodés usar /add para agregar el gasto manualmente.",
	"receipt_no_amount":    "❌ No encontré un monto válido en el recibo. Podés usar /add para agregarlo manualmente.",

	// Examples
	"examples": `📚 *EJEMPLOS DE COMANDOS - BOT DE GASTOS EN PAREJA*

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💰 *AGREGAR GASTOS* (` + "`/add`" + `)

Formato: ` + "`/add <monto> <descripción> [categoría] [método_pago]`" + `

Ejemplos básicos:
• ` + "`/add 50.00 Supermercado`" + `
• ` + "`/add 1250.50 Alquiler`" + `
• ` + "`/add 25.50 Cena Restaurante`" + `

Con categoría:
• ` + "`/add 50.00 Supermercado Comida`" + `
• ` + "`/add 500 Netflix Servicios`" + `

Con método de pago:
• ` + "`/add 50.00 Supermercado Comida Visa`" + `
• ` + "`/add 25.50 Cena Efectivo`" + `

Para tu pareja:
• ` + "`/add 50.00 Supermercado Comida Visa pareja`" + `
• ` + "`/add 25.50 Cena partner`" + `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✏️ *EDITAR GASTOS* (` + "`/edit`" + `)

Formato: ` + "`/edit <id_gasto> <campo> <valor>`" + `

Campos: ` + "`category`" + `, ` + "`payment_method`" + ` (o ` + "`payment`" + `)

Ejemplos:
• ` + "`/edit 123 category Supermercado`" + `
• ` + "`/edit 456 payment_method Visa`" + `
• ` + "`/edit 789 payment Efectivo`" + `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🗑️ *ELIMINAR GASTOS* (` + "`/delete`" + `)

Formato: ` + "`/delete <id_gasto>`" + `

Ejemplos:
• ` + "`/delete 123`" + `
• ` + "`/delete 456`" + `

Sin ID muestra últimos 10 gastos:
• ` + "`/delete`" + `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💳 *AGREGAR MÉTODOS DE PAGO* (` + "`/payment_methods add`" + `)

Formato: ` + "`/payment_methods add <nombre> <tipo> [día_cierre]`" + `

Tipos aceptados (inglés o español):
• ` + "`credit_card`" + ` / ` + "`TarjetaCredito`" + ` / ` + "`tarjeta_credito`" + `
• ` + "`debit_card`" + ` / ` + "`TarjetaDebito`" + ` / ` + "`tarjeta_debito`" + `
• ` + "`cash`" + ` / ` + "`Efectivo`" + `
• ` + "`bank_transfer`" + ` / ` + "`Transferencia`" + `
• ` + "`other`" + ` / ` + "`Otro`" + `

Ejemplos:
• ` + "`/payment_methods add Visa credit_card 15`" + `
• ` + "`/payment_methods add Visa TarjetaCredito 15`" + `
• ` + "`/payment_methods add Debito TarjetaDebito`" + `
• ` + "`/payment_methods add Efectivo cash`" + `
• ` + "`/payment_methods add Transferencia bank_transfer`" + `
• ` + "`/payment_methods add PayPal other`" + `

⚠️ Las tarjetas de crédito requieren día de cierre (1-31)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 *VER MÉTODOS DE PAGO*
• ` + "`/payment_methods`" + `

🔧 *EDITAR MÉTODOS DE PAGO*
• ` + "`/payment_methods edit <id> <campo> <valor>`" + `
• Ejemplo: ` + "`/payment_methods edit 1 closing_day 20`" + `

🗑️ *ELIMINAR MÉTODOS DE PAGO*
• ` + "`/payment_methods delete <id>`" + `
• Ejemplo: ` + "`/payment_methods delete 1`" + `

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

💡 *NOTAS:*
• Montos con punto: ` + "`50.00`" + `, ` + "`1250.50`" + `
• Los IDs se muestran al crear/listar
• Nombres de métodos de pago no distinguen mayúsculas
• Usá ` + "`/help`" + ` para ver todos los comandos`,
}

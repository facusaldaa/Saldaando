# 📚 Ejemplos de Comandos - Bot de Gastos en Pareja

## 💰 Agregar Gastos (`/add`)

### Formato básico:
```
/add <monto> <descripción>
```

### Ejemplos:

**Gasto simple:**
```
/add 50.00 Supermercado
/add 1250.50 Alquiler
/add 3500 Compra de muebles
```

**Con categoría:**
```
/add 50.00 Supermercado Comida
/add 25.50 Cena Restaurante
/add 1200 Alquiler Vivienda
/add 500 Netflix Servicios
```

**Con método de pago:**
```
/add 50.00 Supermercado Comida Visa
/add 25.50 Cena Restaurante Efectivo
/add 1200 Alquiler Vivienda Transferencia
```

**Con categoría y método de pago:**
```
/add 50.00 Supermercado Comida Visa
/add 25.50 Cena Restaurante Efectivo
/add 1200 Alquiler Vivienda Transferencia
```

**Para tu pareja (usando "pareja" o "partner"):**
```
/add 50.00 Supermercado Comida Visa pareja
/add 25.50 Cena Restaurante Efectivo partner
```

**Para usuario específico:**
```
/add 50.00 Supermercado Comida Visa user1
/add 25.50 Cena Restaurante Efectivo user2
```

---

## ✏️ Editar Gastos (`/edit`)

### Formato:
```
/edit <id_gasto> <campo> <valor>
```

### Campos disponibles:
- `category` - Cambiar la categoría
- `payment_method` o `payment` - Cambiar el método de pago

### Ejemplos:

**Cambiar categoría:**
```
/edit 123 category Supermercado
/edit 456 category Restaurante
/edit 789 category Transporte
```

**Cambiar método de pago:**
```
/edit 123 payment_method Visa
/edit 456 payment_method Efectivo
/edit 789 payment_method Transferencia
/edit 123 payment Visa
```

**Ejemplos completos:**
```
/edit 5 category Comida
/edit 10 payment_method TarjetaCredito
/edit 15 category Servicios
/edit 20 payment Efectivo
```

---

## 🗑️ Eliminar Gastos (`/delete`)

### Formato:
```
/delete <id_gasto>
```

### Ejemplos:

**Eliminar por ID:**
```
/delete 123
/delete 456
/delete 789
```

**Ver gastos recientes (sin ID):**
```
/delete
```
*Esto muestra los últimos 10 gastos para que puedas elegir cuál eliminar*

---

## 💳 Agregar Métodos de Pago (`/payment_methods add`)

### Formato:
```
/payment_methods add <nombre> <tipo> [día_cierre]
```

### Tipos disponibles (en inglés o español):

**En inglés:**
- `credit_card` - Tarjeta de crédito
- `debit_card` - Tarjeta de débito
- `cash` - Efectivo
- `bank_transfer` - Transferencia bancaria
- `other` - Otro

**En español (también aceptados):**
- `TarjetaCredito` o `tarjeta_credito` - Tarjeta de crédito
- `TarjetaDebito` o `tarjeta_debito` - Tarjeta de débito
- `Efectivo` - Efectivo
- `Transferencia` o `transferencia_bancaria` - Transferencia bancaria
- `Otro` - Otro

### Ejemplos:

**Tarjeta de crédito (requiere día de cierre):**
```
/payment_methods add Visa credit_card 15
/payment_methods add Mastercard TarjetaCredito 20
/payment_methods add Amex credit_card 5
/payment_methods add Visa tarjeta_credito 15
```

**Tarjeta de débito:**
```
/payment_methods add Debito debit_card
/payment_methods add TarjetaDebito TarjetaDebito
/payment_methods add Débito tarjeta_debito
```

**Efectivo:**
```
/payment_methods add Efectivo cash
/payment_methods add Cash Efectivo
/payment_methods add Dinero efectivo
```

**Transferencia bancaria:**
```
/payment_methods add Transferencia bank_transfer
/payment_methods add Banco Transferencia
/payment_methods add Cuenta bancaria transferencia_bancaria
```

**Otro método:**
```
/payment_methods add PayPal other
/payment_methods add MercadoPago Otro
/payment_methods add Billetera other
```

---

## 📋 Ver Métodos de Pago

```
/payment_methods
```

---

## 🔧 Editar Métodos de Pago

### Formato:
```
/payment_methods edit <id> <campo> <valor>
```

### Ejemplos:
```
/payment_methods edit 1 closing_day 20
/payment_methods edit 2 name Visa Gold
/payment_methods edit 3 type credit_card
/payment_methods edit 4 active false
```

---

## 🗑️ Eliminar Métodos de Pago

### Formato:
```
/payment_methods delete <id>
```

### Ejemplos:
```
/payment_methods delete 1
/payment_methods delete 2
```

---

## 📝 Notas Importantes

1. **Montos**: Usá punto (.) para decimales. Ejemplo: `50.00`, `1250.50`

2. **IDs**: Los IDs de gastos y métodos de pago se muestran cuando los creás o listás

3. **Métodos de pago en gastos**: El nombre del método de pago debe coincidir exactamente con el que creaste (no distingue mayúsculas/minúsculas)

4. **Tarjetas de crédito**: Siempre requieren un día de cierre (1-31)

5. **Español e inglés**: Podés mezclar comandos en español e inglés, ambos funcionan

6. **Ver ayuda completa**: Usá `/help` para ver todos los comandos disponibles

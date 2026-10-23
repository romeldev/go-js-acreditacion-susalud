# Notas técnicas — librería Go X12N SUSALUD

Documento de contexto para quien retome este proyecto. Recoge **qué es la
librería, de dónde sale cada dato del formato, qué defectos tenía el original,
cómo está verificado y qué falta**. Está escrito para que una revisión futura no
tenga que repetir el trabajo de ingeniería inversa.

Si sólo vas a leer una sección, lee [Reglas de mapeo](#7-reglas-de-mapeo-la-fuente-de-bugs)
y [Verificación](#9-verificación). Son las dos que evitan romper cosas.

---

## 1. Resumen

Esta librería genera y parsea los mensajes **X12N / ANSI ASC X12** del juego de
transacciones de **SUSALUD** (Superintendencia Nacional de Salud, Perú):
**270, 271, 278 y 997**, en sus 18 variantes.

Es un port de la librería Java `pe.gob.susalud:jr-acreditacion-susalud-logic:1.0.0`.
El objetivo de aceptación es **paridad byte a byte** con esa librería en la
salida, y lectura tolerante a la entrada.

Estado actual: **18/18 transacciones**, `Marshal` y `Unmarshal`, con verificación
en tres niveles y cero omisiones.

| Métrica | Valor |
|---|---|
| Transacciones | 18 |
| Segmentos X12 definidos | 28 |
| Elementos definidos | 147 |
| Beans (structs de datos) | 31 |
| Campos de bean | 701 |
| Pruebas | 54 subpruebas (18 × 3) |

---

## 2. El artefacto de referencia

La librería original se distribuye como `jr-acreditacion-susalud-1.0.0.dll`.

**No es un DLL de C#.** Es un ensamblado **IKVM**: la compilación a .NET de un
JAR de Java. Datos que conviene tener a mano:

| Dato | Valor |
|---|---|
| Formato | PE32 i386, IL-only, `anycpu` |
| SHA-256 | `8da9970b6f6b4b590e5ee1f1cfae072544e5106ad97343f4f8627ce3fabbca77` |
| Runtime CLR | `v2.0.50727` (.NET Framework 2.0) |
| Ensamblado | versión `0.0.0.0`, **sin firma fuerte** |
| MVID | `a748c486-fb71-47ef-af48-b1dc776c38e4` |
| Referencias | `IKVM.Runtime`, `IKVM.OpenJDK.Core`, `IKVM.OpenJDK.Text` **7.2.4630.5**, `mscorlib` 2.0.0.0 |
| Tipos / métodos | 162 / 2 051 |
| Build | Maven 3.6.3, JDK 1.8.0_212, `source/target 1.6` |

El DLL embebe un recurso `jr-acreditacion-susalud-1.0.0.jar` de **3 922 bytes
que sólo contiene `META-INF/`** — el JAR fue vaciado de `.class` antes de la
conversión. El código real está en los tipos del ensamblado.

**El JAR completo (351 KB, 161 clases) es la fuente de verdad.** Si aparece junto
al DLL, úsalo; si no, se puede extraer el código de los tipos IKVM, pero es mucho
más trabajoso.

### 2.1 Cómo se obtuvo el código fuente

```sh
# decompilar el JAR a Java
java -jar cfr-0.152.jar jr-acreditacion-susalud-1.0.0.jar --outputdir src

# ver el CIL / metadatos del DLL (REA)
#   inspect_managed_artifact  → identidad, ensamblado, recursos
#   inspect_managed_members   → tipos, métodos, campos, literales
```

Se usó **CFR 0.152** (`org.benf:cfr`, descargable de Maven Central). El
decompilado resultante son 22 640 líneas de Java en 161 clases.

> El árbol decompilado vive en `_rea_work/`, que es un directorio de trabajo
> desechable. Todo lo que la librería necesita de él ya está transcrito en
> `edi/segments_gen.go` y `x12n/beans_gen.go`; si borras `_rea_work/` puedes
> seguir trabajando, pero no podrás **regenerar** esos dos archivos.

---

## 3. Formato de línea (lo más importante)

El EDI de SUSALUD **no es X12 estándar**. Es un híbrido:

- Elementos separados por `*`
- Segmentos terminados por `~`
- Componentes (sub-elementos) separados por `:`
- **Y además, cada elemento va rellenado a un ancho fijo**

Las reglas de relleno son las de `trama.Campos.completaLongitud`:

| Clase | Relleno | Ejemplo (`EB01`, ancho 2) |
|---|---|---|
| `Numerico` | `0` a la **izquierda** | `"7"` → `"07"` |
| `Alfanumerico` | espacio a la **derecha** | `"1"` → `"1 "` |
| sin clase | ninguno | `"00501"` queda igual |

Esto se ve en cualquier mensaje real:

```
ISA*00*          *00*          *ZZ*980001C        *ZZ*20001          *171025*1210*|*00501*026483112*0*T*:~
NM1*IL *1*DIAZ      *CLAUDIA LISSET  *  *  *MI*00006726042367*  *  *NARRO
```

`NM1*IL ` — el `IL` de 2 caracteres ocupa un elemento de ancho 3, rellenado con
un espacio. **El receptor depende de ese relleno.** Por eso el encoder lo
reproduce exactamente y no es una opción estética.

### 3.1 El decoder NO usa el ancho fijo

La librería Java tiene un decoder posicional (`SubTrama.procesaSubTrama` que hace
`substring` sobre la concatenación de elementos) — **y nunca se invoca**. Es
código muerto. El parseo real se hace con `split("~")` / `split("*")` + `trim`,
que es lo que hace esta librería también. Consecuencia práctica: **sólo el
encoder es sensible al ancho**. Un `Unmarshal` tolera mensajes sin relleno.

### 3.2 Cabecera de intercambio (constantes)

`SubTrama_ISA` fija estos valores y ninguna transacción los sobrescribe:

| Elemento | Valor | Nota |
|---|---|---|
| ISA05 / ISA07 | `ZZ` | calificador de remitente/receptor |
| ISA09 | `fecha[2:]` | el `yyyyMMdd` recortado a `yyMMdd` |
| ISA10 | `hora[:4]` | `HHmmss` recortado a `HHmm` |
| ISA11 | `\|` | separador de repetición |
| ISA12 | `00501` | versión |
| ISA13 | `IdCorrelativo` | 9 dígitos |
| ISA14 | `0` | sin acuse de intercambio |
| ISA15 | `T` | **siempre test** en el original |
| ISA16 | `:` | separador de componente |

`GS06/GE02` llevan el número de control (8 dígitos). `ST02/SE02` llevan el
control del conjunto de transacción.

`SE01` cuenta los segmentos emitidos **hasta antes del propio `SE`**, que es lo
que hace la referencia (no cumple estrictamente X12, que incluye el `SE`). Se
reproduce tal cual.

---

## 4. Catálogo de segmentos

28 segmentos, 147 elementos. Definidos en `edi/segments_gen.go`, **generado** por
`tools/gen_segments.py` a partir de las clases `trama.SubTrama_*`.

| Segmento | Elementos | Segmento | Elementos |
|---|---|---|---|
| `ISA` | 16 | `EB` | 12 |
| `GS` | 8 | `EB13` | 8 |
| `ST` | 3 | `MSG` | 3 |
| `BHT` | 2 | `HSD` | 2 |
| `HL` | 4 | `N2` | 2 |
| `NM1` | 12 | `N3` | 2 |
| `REF` | 3 | `N4` | 7 |
| `REF4` | 6 | `PER` | 6 |
| `DMG` | 7 | `PRV` | 3 |
| `DTP` | 3 | `INS` | 17 |
| `SE` | 2 | `CRC` | 3 |
| `GE` | 2 | `AK1` | 3 |
| `IEA` | 2 | `AK2` | 3 |
| | | `AK5` | 2 |
| | | `AK9` | 4 |

**`SV1` y `SV2` se excluyeron**: la referencia las declara pero no las instancia
nunca, y `SV2` declara 2 elementos y sólo define 1 (codificar con ella lanzaría
`NullPointerException`).

### 4.1 Regenerar las tablas

```sh
python3 -I tools/gen_segments.py   # → edi/segments_gen.go
```

Lee `_rea_work/src/pe/gob/susalud/jr/transaccion/susalud/trama/SubTrama_*.java`.
**No edites `segments_gen.go` a mano**: el siguiente `go generate` lo pisa, y
además perderías la trazabilidad al original, que es lo que hace auditable el
formato.

---

## 5. Catálogo de transacciones

| Bean | GS01 | BHT02 | Segmentos típicos | Notas |
|---|---|---|---|---|
| `SolAut271` | `HB` | `0023` | 98 | la más compleja del 271 |
| `RegAfi271` | `HB` | `0024` | 137 | la más grande del conjunto |
| `ConAse270` | `HS` | `0022` | 27 | |
| `RegAfi270` | `HS` | `0022` | 14 | |
| `ConNom271` | `HB` | `0022` | 32 | |
| `ConCod271` | `HB` | `0022` | 40 | emite `NM1*3` para el paciente |
| `ConMed271` | `HB` | `0022` | 24 | |
| `In271ConProc` | `HI` | `0022` | 85 | |
| `In271ConObs` | `HB` | `0022` | 16 | |
| `In271ConDtad` | `HB` | `0022` | 18 | usa `N3`/`N4`/`PER` |
| `In271ResSctr` | `HB` | `0022` | 24 | |
| `In271ResDeriva` | `HI` | `0022` | 45 | |
| `LogAcreInsert271` | `HB` | `0022` | 33 | |
| `ConEntVinc278` | `HB` | `0020` | 10 | |
| `ResEntVinc278` | `HB` | `0020` | 10 | |
| `In278SolCG` | `HI` | `0020` | 12 | |
| `In278ResCG` | `HI` | `0022` | 24 | |
| `In997ResAut` | `FA` | — | 10 | sólo transacción sin `BHT` |

`SolAut271` es la única que usa `ST03` con la forma `pe#te#ec#re` (conteo de
detalles por bloque). El resto emite `ST03` vacío.

### 5.1 Bloques con placeholder

Nueve transacciones emiten un **bloque de relleno cuando la lista de detalles
está vacía**, porque el receptor posicional lo necesita:

`SolAut271` (PE, TE, EC y restricciones), `ConMed271`, `ConNom271`, `ConCod271`,
`In278ResCG`, `In271ResSctr`, `In271ConProc`, `In271ResDeriva`, `RegAfi271`.

Al parsear, esos bloques producen un ítem con todos los campos vacíos. **El
parser los descarta** (`prune` + `isZero()` en `x12n/zero.go`) para que
re-serializar vuelva a emitir el placeholder y no un ítem real en blanco. Si
agregas una transacción con placeholder, replica ese patrón o el round-trip
fallará.

---

## 6. Arquitectura Go

```
edi/                    codec, sin conocimiento de negocio
  segment.go            Kind, FieldDef, SegmentDef, Segment, pad, ParseSegment
  segments_gen.go       28 tablas — GENERADO
  envelope.go           ISA, NewControlNumber, Now
  errors.go             FieldLengthError, ParseError

x12n/                   transacciones
  beans_gen.go          31 structs — GENERADO
  encode.go             encoder, Options, sobre ISA/GS/ST…SE/GE/IEA
  <18 archivos>.go      un mapper + un parser por transacción
  zero.go               isZero() por tipo de detalle
  golden_test.go        las 3 suites
  registry_test.go      alta de cada transacción (init)
  testdata/golden.json  fixtures reales
```

**Sin estado de paquete.** `edi.Segment` se crea por mensaje; `x12n` no tiene
variables mutables a nivel de paquete. El parser Java guarda sus contadores en
campos `static`, así que no es reentrante; aquí viajan en una struct por llamada.

### 6.1 API

```go
edi, err := x12n.Marshal271(&bean, &x12n.Options{Production: true})
bean, err := x12n.Unmarshal271(edi)
```

`Options`:

| Campo | Efecto |
|---|---|
| `Control` | ISA13/GS06/GE02. Vacío → se genera (8 dígitos, `crypto/rand`) |
| `ControlST` | ST02/SE02. Vacío → se genera |
| `Production` | ISA15 = `"P"`. Por defecto `"T"`, como el original |
| `Compat` | ver §8.1 |

Los números de control se escriben de vuelta en el bean para poder registrarlos.

---

## 7. Reglas de mapeo (la fuente de bugs)

Las sobrecargas de `generaSubTrama` de la referencia **no escriben en elementos
consecutivos**. Cinco casos son contraintuitivos y cada uno produjo un bug real
durante el port:

| Segmento | Sobrecarga | Elementos que escribe |
|---|---|---|
| `REF4` | 3 args | **1, 2, 4** |
| `REF4` | 2 args | 1, 2 |
| `REF4` | 6 args | 1…6 |
| `EB` | 8 args | **1, 2, 3, 5, 7, 8, 9, 10** |
| `N4` | 2 args | **5, 6** |
| `INS` | 3 args | **1, 2, 17** |
| `PER` | 5 args | **1, 3, 4, 5, 6** |
| `PER` | 2 args | 1, 2 |
| `HL` | 3 args | **1, 3, 4** |
| `HL` | 4 args | 1…4 |
| `NM1` | 2 args | **1, 5** |
| `NM1` | 4 args | **1, 2, 8, 9** |
| `NM1` | 7 args | **1, 2, 3, 4, 8, 9, 12** |
| `GS` | 6 args | 1…6 |
| `ST` | 3 args | 1, 2, 3 |
| `DMG` | 5 args | **1, 2, 3, 4, 7** |
| `EB13` | 3 args | **1, 2, 5** |

Regla mnemotécnica: **`EB` de 8 argumentos es el que más se equivoca** —
`p1→EB01, p2→EB02, p3→EB03, p4→EB05, p5→EB07, p6→EB08, p7→EB09, p8→EB10`. No es
1…8.

Los `REF4` de 3 argumentos se aislaron en el helper `ref4third` (`x12n/conase270.go`).
Al parsearlos hay que leer el **componente 3**, no el 2.

### 7.1 Composiciones anidadas

Los `REF` y `EB` con sub-estructura se emiten **colgando el hijo del último
elemento del padre**, no como segmento nuevo:

```
REF*DD*1*<80 espacios>*4A:089...:...~
                       ^ elemento 4 — el hijo va aquí
```

En el parser: `composite(el, 4)` para `REF`, `composite(el, 13)` para `EB`.
El índice es **siempre el que sigue al último elemento declarado del padre**.

---

## 8. Hallazgos del original y su estado

Todo lo siguiente se verificó sobre el JAR original: por lectura, por grep sobre
las 161 clases, o ejecutando la librería.

| # | Hallazgo | Verificación | Estado en Go |
|---|---|---|---|
| 1 | `validaFecha` usa `"YYYYmmdd"`: en `SimpleDateFormat` `Y` es semana-año y `m` es minuto. Acepta `20231345`, `20230230`, `00000000` | sonda Java ejecutada | no portado (validadores) |
| 2 | Los validadores **nunca se invocan** desde los `ServiceImpl` | grep | no portado a propósito |
| 3 | `size() >= 0` siempre es true → los campos del afiliado registrado (códigos 0700-0710, 1400-1410) **nunca se validan** | lectura | no portado |
| 4 | Sin escape de `*`, `~` ni `:` en texto libre → un producto con `*` rompe el mensaje | lectura | **no reproducido**: ver pendientes |
| 5 | `ISA15` fijo en `"T"`: todo sale marcado como prueba | lectura | **corregido**: `Options.Production` |
| 6 | 1 solo `try/catch` en 22 640 líneas; los parsers lanzan `ArrayIndexOutOfBounds` con datos reales | grep + ejecución del demo | **corregido**: parser tolerante |
| 7 | `generarAleatorio` hace `new Random()` por llamada; colisión por cumpleaños a ~37k mensajes | lectura | **corregido**: `crypto/rand` |
| 8 | `completaLongitud` **borra** los valores que exceden su ancho (los sustituye por `""`/`"0"`) | lectura + golden | **opcional**: `Options.Compat` |
| 9 | 825 `System.out.println` en una librería | grep | **corregido**: errores por retorno |
| 10 | Estado mutable `static` en los parsers → no reentrante | lectura | **corregido** |
| 11 | `procesaSubTrama` (decoder posicional) nunca se llama: código muerto | grep | omitido |
| 12 | `SV1`/`SV2` declaradas y nunca instanciadas; `SV2` con 2 campos y 1 definido | grep | omitido |
| 13 | `In271ResSctrServiceImpl` rompe la interfaz: expone `In271ResSctr_ToX12N` en vez de `beanToX12N` | lectura | **corregido**: API uniforme |
| 14 | Ensamblado sin firma, versión `0.0.0.0` | metadatos PE | N/A |
| 15 | El JAR embebido en el DLL está vaciado (sólo `META-INF`) | extracción + hash | N/A |
| 16 | `cantidaSegmentosTx` devuelve `null` si hay menos de 2 segmentos → NPE | lectura | **corregido** |
| 17 | El conteo `ST03` con 0 se sustituye por 1 antes de leerlo | lectura | reproducido a propósito |

### 8.1 `Options.Compat` en detalle

El origen de esta opción es el hallazgo #8. La referencia, ante un valor más
largo que su ancho, **no trunca ni falla: borra**. Por ejemplo `DeflagCG`
(18 caracteres) se asigna a `EB04`, que mide 3, y el campo sale vacío.

Por defecto esta librería devuelve `*edi.FieldLengthError` — un dato que no cabe
es un error del llamador y debe verse.

`Compat: true` reproduce el borrado. **Sólo se usa en los golden tests**, para
demostrar paridad byte a byte incluso en ese caso. No lo actives en producción
sin una razón: reproduce pérdida silenciosa de datos.

---

## 9. Verificación

Tres suites, todas contra la **salida real** de la librería Java. Ninguna
aserción se basa en lectura del código.

| Suite | Recorrido | Qué demuestra |
|---|---|---|
| `TestGolden` | bean → EDI | byte a byte idéntico al Java |
| `TestRoundTrip` | EDI → bean → EDI | el parser no pierde nada |
| `TestBeanRoundTrip` | bean ↔ bean | el bean parseado es idéntico campo a campo |

La tercera es la que atrapa un mapeo invertido: si el encoder y el decoder
cometieran el mismo error de elemento, el round-trip pasaría igual.

### 9.1 Los fixtures

`x12n/testdata/golden.json` contiene, por transacción, el bean de muestra que
construye `TransaccionDemo` y el EDI que produce la librería a partir de él.

Se generan volcando el bean **por reflexión** después de llamar al servicio:

```sh
javac -cp <jar> -d tools/golden tools/golden/BeanDump.java
java -cp tools/golden:<jar> BeanDump x12n/testdata/golden.json
```

Dos detalles del arnés que cuestan tiempo si se redescubren:

- La librería imprime a `stdout`, así que el JSON **se escribe a un archivo**,
  no a la salida estándar.
- Los números de control se leen **del bean**, no de `GS06`/`ST02`: en el wire
  van rellenos con ceros y el valor original es irrecuperable.

### 9.2 Ejecutar

```sh
go test ./...                          # todo
go test ./x12n/ -run 'TestGolden/SolAut271' -v
```

Ojo: `-run` con un solo patrón no entra en subtests. Usa
`-run 'Test(Golden|RoundTrip|BeanRoundTrip)/<Txn>'`.

---

## 10. Receta: agregar o cambiar una transacción

1. **Localiza el traductor** en el Java decompilado:
   `_rea_work/src/.../util/<Bean>_ToX12N.java` → `traducirEstructura*`.
2. **Anota el orden de segmentos** del `sReturn` y el mapeo de cada
   `generaSubTrama`. Consulta la tabla de §7 para cada sobrecarga.
3. **Escribe el `Marshal`** en `x12n/<bean>.go`, usando
   `e.open(...)` / `e.close(...)` para el sobre y `ref`, `ref4`, `ref4third`,
   `eb13`, `nm1Person` para los patrones comunes.
4. **Registra la transacción** en `x12n/registry_test.go` (mapa `marshals`) y
   corre `TestGolden`. El fallo te dice **el número de segmento exacto** y el
   valor esperado contra el obtenido — itera hasta que pase.
5. **Escribe el `Unmarshal`** replicando los contadores de la referencia
   (arrancan en **1**, no en 0) y las reglas de avance condicional en bloques
   repetidos.
6. **Si hay lista de detalles**, añade `isZero()` en `x12n/zero.go` y llama a
   `prune`. Sin esto, el placeholder de lista vacía rompe el round-trip.
7. **Registra** en `roundtrips` y `beanRoundTrips`, y corre las tres suites.
8. **Actualiza** la tabla de §5 de este documento y el estado del README.

Verifica siempre: `go vet ./... && go test ./...`.

### 10.1 Señales de que algo está mal

| Síntoma | Causa probable |
|---|---|
| Un solo elemento desplazado, todo lo demás bien | sobrecarga mal mapeada (§7) |
| El round-trip falla justo en un placeholder | falta `prune`/`isZero` |
| El bean difiere en 1-3 campos pero el EDI coincide | índice de componente anidado (§7.1) |
| `FieldLengthError` en el golden | dato de muestra que excede su ancho → usa `Compat` en el test |
| Un contador no avanza / avanza de más | bloque repetido: el avance es condicional |

---

## 11. Límites conocidos y trabajo pendiente

### 11.1 No portado, a propósito

- **Capa de validación de negocio** (`validator/` en el Java: 18 clases,
  ~3 000 líneas). No se portó porque en el original **nunca se ejecuta** y, si se
  ejecutara, está rota (hallazgos #1 y #3). Antes de escribirla hay que decidir
  qué reglas se quieren de verdad — no tiene sentido transcribir 3 000 líneas de
  validación inerte y defectuosa.
- **`procesaSubTrama`** (decoder posicional): código muerto en el original.

### 11.2 Riesgo latente no resuelto

**Los delimitadores no se escapan.** Si un campo de texto libre
(`DeProducto`, `ObsRestricciones`, `MsgRestricciones`, …) contiene `*` o `~`, el
mensaje se corrompe y todos los segmentos siguientes se desplazan. El original
tampoco lo resuelve: tiene un `removerCaractEspecial` que **nunca se llama** y que
además no elimina esos caracteres.

Opciones para cuando se aborde:

1. Rechazar el valor en `Marshal` con un error tipado (rompe compatibilidad pero
   es seguro).
2. Sanear sustituyendo por un carácter seguro (cambia datos en silencio).
3. Dejarlo al llamador y documentarlo.

No hay decisión tomada. **Es el punto más importante a resolver antes de exponer
la librería a datos externos.**

### 11.3 Otros

- Sin verificación de totales de control (`SE01`, `GE01`, `IEA01`): se emiten
  como los emite el original, sin recalcular ni comparar al leer.
- Sin `TA1` ni emisión de `997` desde las transacciones (el `997` existe como
  transacción propia, pero ningún servicio lo genera en respuesta).

---

## 12. Decisiones abiertas

Preguntas que una revisión futura debería resolver con el usuario:

1. **Escape de delimitadores** (§11.2): ¿rechazar, sanear o documentar?
2. **Validación de negocio**: ¿se implementa? ¿con qué reglas? ¿devolviendo
   códigos de error como el original (`0100`, `0101`, …) o errores Go?
3. **Modo por defecto**: ¿`Compat` debería pasar a ser el defecto cuando el
   contraparte sea estrictamente la librería vieja?
4. **Publicación**: el módulo ya es `github.com/romeldev/go-js-acreditacion-susalud`.
   Falta fijar la licencia — sin archivo `LICENSE` el repo queda como "todos los
   derechos reservados" por defecto, lo que desalienta su uso.
5. **`SE01`**: se reproduce el conteo no estándar del original. Si algún día se
   integra con un validador X12 externo, esto habrá que corregirlo (y
   posiblemente romper paridad de bytes).

---

## 13. Referencias rápidas

| Quiero… | Archivo |
|---|---|
| Ver el formato de un segmento | `edi/segments_gen.go` |
| Ver qué campos tiene un bean | `x12n/beans_gen.go` |
| Cambiar el sobre ISA/GS/ST | `x12n/encode.go` |
| Ver el patrón de una transacción simple | `x12n/in997.go` |
| Ver el patrón complejo (bloques repetidos) | `x12n/solaut271.go` + `solaut271_parse.go` |
| Añadir una transacción | `x12n/registry_test.go` |
| Diferencias deliberadas con el original | `README.md` |
| Generar tablas desde el Java | `tools/gen_segments.py`, `tools/gen_beans.py` |
| Generar fixtures desde el JAR | `tools/golden/BeanDump.java` |

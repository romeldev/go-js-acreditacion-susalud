# susalud — librería Go de transacciones X12N SUSALUD

Port a Go de la librería Java `pe.gob.susalud:jr-acreditacion-susalud-logic:1.0.0`
(la que se distribuye compilada como `jr-acreditacion-susalud-1.0.0.dll` vía IKVM).

Genera y parsea los mensajes **X12N / ANSI ASC X12** del juego de transacciones
SUSALUD: **270, 271, 278 y 997**, en sus 18 variantes.

La salida es **byte a byte idéntica** a la de la implementación Java, incluidas
sus particularidades de formato. Los defectos de la original *no* se copian por
defecto; ver [Diferencias deliberadas](#diferencias-deliberadas).

## Estado

**Las 18 transacciones están completas**: `Marshal` y `Unmarshal`, con EDI
byte-idéntico al de la librería Java, round-trip exacto y bean parseado idéntico
campo a campo. 54 subpruebas, ninguna omitida.

| Transacción | Segmentos (referencia) |
|---|---|
| SolAut271 | 98 |
| RegAfi271 | 137 |
| In271ConProc | 85 |
| In271ResDeriva | 45 |
| RegAfi270 | 14 |
| ConAse270 | 27 |
| ConNom271 | 32 |
| ConCod271 | 40 |
| ConMed271 | 24 |
| ConEntVinc278 | 10 |
| In278SolCG | 12 |
| In278ResCG | 24 |
| ResEntVinc278 | 10 |
| In271ResSctr | 24 |
| In271ConObs | 16 |
| In271ConDtad | 18 |
| LogAcreInsert271 | 33 |
| In997ResAut | 10 |

Cada transacción portada se verifica en tres niveles contra la salida real de la
librería Java (fixtures en `x12n/testdata/golden.json`):

1. **EDI idéntico** — `TestGolden`: el bean de muestra se re-serializa y debe
   coincidir byte a byte con el que produjo el Java.
2. **Round-trip** — `TestRoundTrip`: el EDI del Java se parsea y se vuelve a
   serializar; debe reproducirse exactamente.
3. **Bean idéntico** — `TestBeanRoundTrip`: el bean parseado se compara campo a
   campo con el bean original. Este nivel atrapa un mapeo invertido que el
   round-trip por sí solo no vería (si el mismo error estuviera en ambos lados).

Una transacción sin portar se marca como `SKIP` en `go test`.

## Estructura

```
edi/                  codec de bajo nivel
  segment.go          SegmentDef, Segment, padding, ParseSegment
  segments_gen.go     las 28 tablas de segmento (generado)
  envelope.go         ISA, números de control
  errors.go
x12n/                 transacciones
  beans_gen.go        31 structs de datos (generado)
  encode.go           sobre ISA/GS/ST/SE/GE/IEA + política de desborde
  solaut271.go        271
  in997.go            997
  marshal278.go       278 (enlace de entidades, respuesta, solicitud CG)
  golden_test.go      verificación contra la referencia
docs/
  NOTAS-TECNICAS.md   contexto para futuras revisiones: formato, mapeos,
                      hallazgos del original, recetas de mantenimiento
tools/
  gen_segments.py     extrae las tablas de segmento del Java decompilado
  gen_beans.py        extrae el modelo de datos
  golden/             arnés Java que produce los fixtures
```

> **Antes de modificar el formato o portar una transacción nueva, lee
> [`docs/NOTAS-TECNICAS.md`](docs/NOTAS-TECNICAS.md).** Documenta las 17
> sobrecargas de `generaSubTrama` con los elementos que realmente escriben —
> la fuente de casi todos los bugs de este port.

## Uso

```go
b := &x12n.InSolAut271{
    IdRemitente: "980001C",
    IdReceptor:  "20001",
    FeTransaccion: "20261022",
    HoTransaccion: "121037",
    IdCorrelativo: "026483112",
    IdTransaccion: "271",
    // ...
}

edi, err := x12n.Marshal271(b, &x12n.Options{Production: true})
```

Los números de control se generan con `crypto/rand` si `Options.Control` /
`Options.ControlST` están vacíos, y se escriben de vuelta en el bean para que
puedan registrarse.

Leer un mensaje entrante:

```go
b, err := x12n.Unmarshal271(cuerpo)
if err != nil {
    // mensaje no interpretable
}
log.Printf("correlativo %s, paciente %s", b.IdCorrelativo, b.NoPaciente)
```

El parser es tolerante: un elemento ausente al final del segmento se lee como
`""` en vez de provocar `ArrayIndexOutOfBoundsException` como hace el original.
Los números de control (`GS06`/`ST02`) no se restauran, porque el EDI los lleva
rellenos con ceros y el valor original no es recuperable; son parámetros de
salida, no datos de negocio. Si necesitas reproducir un intercambio concreto,
pásalos por `Options`.

## Formato de línea

El EDI es híbrido: elementos separados por `*`, segmentos por `~`, y **cada
elemento rellenado a su ancho declarado** (numérico con `0` a la izquierda,
alfanumérico con espacios a la derecha). Ese relleno es el contrato real: el
parser del contraparte lo espera.

```
ISA*00*          *00*          *ZZ*980001C        *ZZ*20001          *171025*1210*|*00501*026483112*0*T*:~
NM1*IL *1*DIAZ      *CLAUDIA LISSET  *  *  *MI*00006726042367*  *  *NARRO
```

Los anchos no están escritos a mano: `tools/gen_segments.py` los transcribe de
las clases `trama.SubTrama_*` decompiladas, de modo que ninguna longitud se
reinterpreta.

## Diferencias deliberadas

La implementación Java tiene defectos que se corrigen aquí. La salida sigue
siendo idéntica para entradas válidas.

| Java | Go |
|---|---|
| Un elemento que excede su ancho se **borra en silencio** (se reemplaza por `""`/`"0"`) | Devolver `*edi.FieldLengthError`. `Options.Compat` reproduce el comportamiento original si un contraparte lo exige |
| Número de control con `new Random()` por llamada; colisions por cumpleaños a ~37k mensajes | `crypto/rand`, uniforme sobre el rango completo |
| `state` en campos `static`: el parser no es seguro para concurrencia | Sin estado de paquete; uso concurrente seguro |
| Errores por `System.out.println` (825 llamadas) | Errores tipados por retorno |
| Validadores nunca invocados desde los servicios | (pendiente: exponer `Validate` por transacción) |
| `ISA15` fijo en `"T"` (siempre marcas de prueba) | `Options.Production` → `"P"` |
| `substring` sin guardas (`ISA09`, `ISA10`) | Error tipado si fecha/hora son cortas |

Las tablas de segmento y el modelo de datos son **generados**; no editar
`edi/segments_gen.go` ni `x12n/beans_gen.go` a mano.

## Verificación

`x12n/testdata/golden.json` contiene, por transacción, el bean de muestra que
construye `TransaccionDemo` y el EDI que la librería Java produce a partir de
él. `TestGolden` repuebla el struct Go con ese bean, vuelve a serializar y exige
igualdad byte a byte.

Para regenerar los fixtures contra el JAR original:

```sh
javac -cp <jar> -d tools/golden tools/golden/BeanDump.java
java -cp tools/golden:<jar> BeanDump x12n/testdata/golden.json
```

## Licencia

MIT — ver [LICENSE](LICENSE). Copyright (c) 2026 Romel Diaz.

Este repositorio contiene **sólo** el port a Go. La implementación Java de
referencia y sus binarios no se distribuyen aquí; el árbol decompilado que se
usó para extraer el formato está excluido por `.gitignore`.

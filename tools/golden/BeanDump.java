import java.lang.reflect.*;
import java.util.*;

/**
 * Dumps, for every transaction in the reference library, the sample bean built
 * by TransaccionDemo together with the EDI the reference implementation
 * produces from it. The output is one JSON object per transaction:
 *
 *   {"txn":"SolAut271","impl":"SolAut271ServiceImpl",
 *    "edi":"ISA*...", "bean":{...}}
 *
 * It is the fixture the Go port is verified against.
 *
 * Usage: java -cp <out>:<jar> BeanDump > golden.json
 */
public class BeanDump {

    private static final String[][] MAP = {
        {"ConAse270",      "generarConAse270CC",      "ConAse270ServiceImpl"},
        {"RegAfi270",      "generarRegafi270",        "RegAfi270ServiceImpl"},
        {"RegAfi271",      "generarRegafi271",        "RegAfi271ServiceImpl"},
        {"ConNom271",      "generarConNom271",        "ConNom271ServiceImpl"},
        {"ConCod271",      "generarConCod271",        "ConCod271ServiceImpl"},
        {"ConMed271",      "generarConMed271",        "ConMed271ServiceImpl"},
        {"In278SolCG",     "generarIn278Sol",         "In278SolCGServiceImpl"},
        {"In278ResCG",     "generarIn278Res",         "In278ResCGServiceImpl"},
        {"In271ResSctr",   "generarIn271ResSctr",     "In271ResSctrServiceImpl"},
        {"SolAut271",      "generarSolAut271",        "SolAut271ServiceImpl"},
        {"In271ConProc",   "generarIn271ConProc",     "In271ConProcServiceImpl"},
        {"In997ResAut",    "generarIn997ResAut",      "In997ResAutServiceImpl"},
        {"In271ConObs",    "generarIn271ConObs",      "In271ConObsServiceImpl"},
        {"ConEntVinc278",  "generarInSolEntVinc278",  "ConEntVinc278ServiceImpl"},
        {"ResEntVinc278",  "generarInResEntVinc278",  "ResEntVinc278ServiceImpl"},
        {"LogAcreInsert271","generarLogAcreInsert271","LogAcreInsert271ServiceImpl"},
        {"In271ConDtad",   "generarIn271ConDtad",     "In271ConDtadServiceImpl"},
        {"In271ResDeriva", "generarIn271ResDeriva",   "In271ResDerivaServiceImpl"},
    };

    public static void main(String[] args) throws Exception {
        Class<?> demo = Class.forName(
            "pe.gob.susalud.jr.transaccion.susalud.demo.TransaccionDemo");

        StringBuilder out = new StringBuilder();
        out.append("[\n");
        boolean first = true;

        for (String[] entry : MAP) {
            String txn = entry[0], gen = entry[1], impl = entry[2];
            Method g = demo.getDeclaredMethod(gen);
            g.setAccessible(true);
            Object bean = g.invoke(null);

            Class<?> svc = Class.forName(
                "pe.gob.susalud.jr.transaccion.susalud.service.imp." + impl);
            Object instance = svc.getDeclaredConstructor().newInstance();

            Method b2x = null;
            for (Method m : svc.getMethods()) {
                if (m.getParameterCount() == 1
                        && m.getParameterTypes()[0] == bean.getClass()) {
                    b2x = m;
                    break;
                }
            }
            String edi = (String) b2x.invoke(instance, bean);

            if (!first) out.append(",\n");
            first = false;
            out.append("{\"txn\":").append(str(txn));
            out.append(",\"impl\":").append(str(impl));
            out.append(",\"edi\":").append(str(edi));
            out.append(",\"bean\":");
            out.append(dump(bean, 0));
            out.append("}");
        }
        out.append("\n]\n");
        // The reference library prints diagnostics to stdout, so the fixture is
        // written to the path given as the first argument instead.
        if (args.length > 0) {
            java.nio.file.Files.write(
                java.nio.file.Paths.get(args[0]),
                out.toString().getBytes(java.nio.charset.StandardCharsets.UTF_8));
        } else {
            System.out.print(out);
        }
    }

    private static String dump(Object o, int depth) throws Exception {
        if (o == null) return "null";
        if (o instanceof String) return str((String) o);
        if (o instanceof Number || o instanceof Boolean) return o.toString();
        if (o instanceof Collection) {
            StringBuilder sb = new StringBuilder("[");
            boolean first = true;
            for (Object item : (Collection<?>) o) {
                if (!first) sb.append(",");
                first = false;
                sb.append(dump(item, depth + 1));
            }
            return sb.append("]").toString();
        }
        StringBuilder sb = new StringBuilder("{");
        boolean first = true;
        for (Field f : o.getClass().getDeclaredFields()) {
            if (Modifier.isStatic(f.getModifiers())) continue;
            f.setAccessible(true);
            Object v = f.get(o);
            if (v == null) continue;
            if (v instanceof String && ((String) v).isEmpty()) continue;
            if (!first) sb.append(",");
            first = false;
            sb.append(str(f.getName())).append(":").append(dump(v, depth + 1));
        }
        return sb.append("}").toString();
    }

    private static String str(String s) {
        StringBuilder sb = new StringBuilder("\"");
        for (int i = 0; i < s.length(); i++) {
            char c = s.charAt(i);
            switch (c) {
                case '"':  sb.append("\\\""); break;
                case '\\': sb.append("\\\\"); break;
                case '\n': sb.append("\\n"); break;
                case '\r': sb.append("\\r"); break;
                case '\t': sb.append("\\t"); break;
                default:
                    if (c < 0x20) sb.append(String.format("\\u%04x", (int) c));
                    else sb.append(c);
            }
        }
        return sb.append("\"").toString();
    }
}

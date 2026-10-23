import java.lang.reflect.*;

/**
 * Generates reference EDI output for every transaction in the reference
 * library, reusing the sample beans built by TransaccionDemo's private
 * generators. Output is one block per transaction, delimited by "### name".
 *
 * Usage: java -cp <out>:<jar> GoldenGen > goldens.txt
 */
public class GoldenGen {

    private static final String[][] MAP = {
        {"generarConAse270CC",      "ConAse270ServiceImpl"},
        {"generarRegafi270",        "RegAfi270ServiceImpl"},
        {"generarRegafi271",        "RegAfi271ServiceImpl"},
        {"generarConNom271",        "ConNom271ServiceImpl"},
        {"generarConCod271",        "ConCod271ServiceImpl"},
        {"generarConMed271",        "ConMed271ServiceImpl"},
        {"generarIn278Sol",         "In278SolCGServiceImpl"},
        {"generarIn278Res",         "In278ResCGServiceImpl"},
        {"generarIn271ResSctr",     "In271ResSctrServiceImpl"},
        {"generarSolAut271",        "SolAut271ServiceImpl"},
        {"generarIn271ConProc",     "In271ConProcServiceImpl"},
        {"generarIn997ResAut",      "In997ResAutServiceImpl"},
        {"generarIn271ConObs",      "In271ConObsServiceImpl"},
        {"generarInSolEntVinc278",  "ConEntVinc278ServiceImpl"},
        {"generarInResEntVinc278",  "ResEntVinc278ServiceImpl"},
        {"generarLogAcreInsert271", "LogAcreInsert271ServiceImpl"},
        {"generarIn271ConDtad",     "In271ConDtadServiceImpl"},
        {"generarIn271ResDeriva",   "In271ResDerivaServiceImpl"},
    };

    public static void main(String[] args) throws Exception {
        Class<?> demo = Class.forName(
            "pe.gob.susalud.jr.transaccion.susalud.demo.TransaccionDemo");

        for (String[] entry : MAP) {
            String gen = entry[0], impl = entry[1];
            System.out.println("### " + gen + " " + impl);
            try {
                Method g = demo.getDeclaredMethod(gen);
                g.setAccessible(true);
                Object bean = g.invoke(null);

                Class<?> svc = Class.forName(
                    "pe.gob.susalud.jr.transaccion.susalud.service.imp." + impl);
                Object instance = svc.getDeclaredConstructor().newInstance();

                Method b2x = null;
                try {
                    b2x = svc.getMethod("beanToX12N", bean.getClass());
                } catch (NoSuchMethodException e) {
                    for (Method m : svc.getMethods()) {
                        if (m.getParameterCount() == 1
                                && m.getParameterTypes()[0] == bean.getClass()) {
                            b2x = m;
                            break;
                        }
                    }
                }
                if (b2x == null) {
                    System.out.println("!! no beanToX12N on " + impl);
                    continue;
                }
                String edi = (String) b2x.invoke(instance, bean);
                System.out.println("@@ " + b2x.getName());
                System.out.println(edi);
            } catch (Throwable t) {
                Throwable cause = t instanceof InvocationTargetException
                    ? ((InvocationTargetException) t).getTargetException() : t;
                System.out.println("!! " + cause);
            }
            System.out.println();
        }
    }
}

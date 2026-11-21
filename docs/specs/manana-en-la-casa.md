# Una mañana en la casa (M11)

El escenario `tests/scenarios/una_manana_en_la_casa.json` reúne un dormitorio con cama y escritorio, una cocina y un living con televisor y mesa. La giantess tiene hambre, energía y diversión, y elige entre cinco ofertas. Veinte tinies comparten un nido en el escritorio y pueden recolectar, comer, descansar o explorar según sus rasgos.

El único comando del escenario, en el tick cero, configura las referencias de la giantess a su palma. Preparar el desayuno deja una taza y migajas, y comer mantiene la mano apoyada. La oferta de mirar televisión apoya la mano sobre la mesa del living. Objetos, necesidades, percepción, trabajos, poses y portales usan los sistemas del motor.

La aceptación correlaciona eventos por entidad y destino. Exige que el desayuno deje recursos en el escritorio, que la colonia los descubra y asigne un acarreo, y que el mismo tiny suba a la palma, viaje con ella cuando la giantess levanta la mano y cambia de habitación, y baje a explorar el living. `TestUnaMananaEnLaCasa` usa las semillas 1, 7, 21, 42 y 99, verifica los invariantes cada tick y compara el log y el hash al continuar desde snapshots JSON tomados a lo largo de cada partida.

El visor abre el escenario por defecto. `H` alterna entre historia y eventos, y `RePág` y `AvPág` recorren el historial. Seleccioná una entidad con un clic y pulsá `T` para seguirla. El seguimiento usa la posición interpolada y mantiene la vista alineada con el mundo. `F` vuelve a centrar la entidad seguida y `Esc` libera la cámara.

El benchmark mide todos los ticks del escenario, incluida la comprobación de invariantes y la captura del log. Opcionalmente, `-seed` reemplaza la semilla del escenario y `-ticks` cambia la duración.

```sh
go run ./cmd/game
go run ./cmd/simrun -scenario tests/scenarios/una_manana_en_la_casa.json -events
go run ./cmd/simbench -scenario tests/scenarios/una_manana_en_la_casa.json
```

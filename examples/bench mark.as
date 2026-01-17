{
  Console.log("------------------ do-while --------------------");
  var loop = 0;
  #_bench_start()
  do {
    loop++
  } while (loop < 10 ** 3)
  #_bench_end()
  Console.log("looped", loop, "times");
}

{
  Console.log("------------------ while --------------------");
  var loop = 0;
  #_bench_start()
  while (loop < 10 ** 3) {
    loop++
  }
  #_bench_end()
  Console.log("looped", loop, "times");
}

{
  Console.log("------------------ for --------------------");
  var loop = 0;
  #_bench_start()
  for (i = 0; i <= (10 ** 3); i++) {
    loop = i;
  }
  #_bench_end()
  Console.log("looped", loop, "times");
}

spawn object = {};
#_bench_start()
for (i = 0; i < 10 ** 3; i++) {
  object[i] = null;
}
#_print("created object in ");
#_bench_end()

{
  Console.log("------------------ for-in --------------------");
  #_bench_start()
  var loop = 0;
  for (spawn k in object) {
    loop++;
  }
  #_bench_end()
  Console.log("looped", loop, "times");
}

{
  Console.log("------------------ for..of --------------------");
  #_bench_start()
  var loop = 0;
  for (spawn v of object) {
    loop++;
  }
  #_bench_end()
  Console.log("looped", loop, "times");
}

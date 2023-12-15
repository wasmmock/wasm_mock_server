export function deepEqual(object1, object2) {
  const keys1 = Object.keys(object1);
  const keys2 = Object.keys(object2);
  var keys1_1 = []
  var keys2_1 = []
  for (var i=0;i<keys1.length;i++){
    if (object1[keys1[i]]!=null){
      keys1_1.push(keys1[i])
    }
  }
  for (var i=0;i<keys2.length;i++){
    if (object2[keys2[i]]!=null){
      keys2_1.push(keys2[i])
    }
  }
  if (keys1_1.length !== keys2_1.length) {
    return false;
  }

  for (const key of keys2_1) {
    const val1 = object1[key];
    const val2 = object2[key];
    const areObjects = isObject(val1) && isObject(val2);
    if (
      areObjects && !deepEqual(val1, val2) ||
      !areObjects && val1 !== val2
    ) {
      return false;
    }
  }

  return true;
}

function isObject(object) {
  return object != null && typeof object === 'object';
}
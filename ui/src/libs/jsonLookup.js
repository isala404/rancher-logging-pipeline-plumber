/* eslint-disable no-prototype-builtins */
/* eslint-disable no-restricted-syntax */
/* eslint-disable guard-for-in */
export default function keyLookup(obj, k) {
  if (typeof (obj) !== 'object') {
    return null;
  }
  let result = null;
  if (obj.hasOwnProperty(k)) {
    return obj[k];
  }
  for (const o in obj) {
    result = keyLookup(obj[o], k);
    // eslint-disable-next-line no-continue
    if (result == null) continue;
    else break;
  }

  return result;
}

const fs = require('fs');
const orig = fs.cpSync;
fs.cpSync = function(s, d, o) {
  if (o && o.recursive && !o.filter) {
    o = Object.assign({}, o, { filter: () => true });
  }
  return orig.call(this, s, d, o);
};

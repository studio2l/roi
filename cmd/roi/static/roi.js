function rfc3339(d) {
  function pad(n) {
    return n < 10 ? "0" + n : n;
  }
  function timezoneOffset(offset) {
    var sign;
    if (offset === 0) {
      return "Z";
    }
    sign = (offset > 0) ? "-" : "+";
    offset = Math.abs(offset);
    return sign + pad(Math.floor(offset / 60)) + ":" + pad(offset % 60);
  }
  return d.getFullYear() + "-" +
    pad(d.getMonth() + 1) + "-" +
    pad(d.getDate()) + "T" +
    pad(19) + ":" +
    pad(0) + ":" +
    pad(0) +
    timezoneOffset(d.getTimezoneOffset());
}

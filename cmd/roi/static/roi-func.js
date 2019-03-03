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
        pad(23) + ":" +
        pad(59) + ":" +
        pad(59) +
        timezoneOffset(d.getTimezoneOffset());
}

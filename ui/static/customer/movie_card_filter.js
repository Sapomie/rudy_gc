(function () {
    function normalizeValue(element) {
        if (!element) {
            return "";
        }
        return String(element.value || "").trim();
    }

    function shouldSkipElement(element) {
        if (!element || !element.name || element.disabled) {
            return true;
        }
        var type = String(element.type || "").toLowerCase();
        return type === "submit" || type === "button" || type === "reset" || type === "fieldset";
    }

    document.addEventListener("DOMContentLoaded", function () {
        var form = document.getElementById("movieCardFilterForm");
        if (!form) {
            return;
        }

        Array.prototype.forEach.call(form.elements, function (element) {
            if (shouldSkipElement(element)) {
                return;
            }
            element.dataset.initialValue = normalizeValue(element);
        });

        form.addEventListener("submit", function (event) {
            event.preventDefault();

            var params = new URLSearchParams();

            Array.prototype.forEach.call(form.elements, function (element) {
                if (shouldSkipElement(element)) {
                    return;
                }

                var name = element.name;
                var value = normalizeValue(element);
                var type = String(element.type || "").toLowerCase();
                var explicit = element.dataset.explicit === "1";
                var initialValue = element.dataset.initialValue || "";

                if ((type === "checkbox" || type === "radio") && !element.checked) {
                    return;
                }

                if (type === "hidden") {
                    if (explicit && value !== "") {
                        params.set(name, value);
                    }
                    return;
                }

                if (value === "") {
                    return;
                }

                if (explicit || value !== initialValue) {
                    params.set(name, value);
                }
            });

            var action = form.getAttribute("action") || window.location.pathname;
            var query = params.toString();
            window.location.assign(query ? action + "?" + query : action);
        });
    });
})();

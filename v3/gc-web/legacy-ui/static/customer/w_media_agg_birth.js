document.addEventListener("DOMContentLoaded", function () {
    var titleNodes = [].slice.call(document.querySelectorAll('[title]'));
    titleNodes.forEach(function (el) {
        if (el.getAttribute("data-bs-toggle") !== "tooltip") {
            el.setAttribute("data-bs-toggle", "tooltip");
            if (window.bootstrap && bootstrap.Tooltip) {
                new bootstrap.Tooltip(el, {boundary: document.body, container: "body"});
            }
        }
    });

    document.querySelectorAll(".nav.nav-tabs").forEach(function (group) {
        var active = group.querySelector(".nav-link.active");
        if (active) {
            return;
        }
        var first = group.querySelector('[data-bs-toggle="tab"]');
        if (first && window.bootstrap && bootstrap.Tab) {
            new bootstrap.Tab(first).show();
        }
    });

    var toggleBtn = document.getElementById("toggleTopAgg");
    var topCollapse = document.getElementById("topAggCollapse");
    if (toggleBtn && topCollapse) {
        topCollapse.addEventListener("shown.bs.collapse", function () {
            toggleBtn.textContent = "收起";
        });
        topCollapse.addEventListener("hide.bs.collapse", function () {
            toggleBtn.textContent = "展开";
        });
    }
});
